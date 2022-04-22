package merge

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"github.com/inngest/event-schemas/pkg/cueutil"
)

// Merge merges two cue Values, returning a fully formatted cue Value which represents
// the merged definition.
func Merge(ctx context.Context, a, b cue.Value) (cue.Value, error) {
	merged, err := recursivelyMerge(ctx, a, b)
	if err != nil {
		return merged, err
	}

	merged, err = recursivelyMerge(ctx, b, merged)
	if err != nil {
		return merged, err
	}

	return merged, nil
}

func recursivelyMerge(ctx context.Context, a, b cue.Value) (cue.Value, error) {
	// If one of the values is BottomKind it has no data, so we can return
	// the other value immediately.
	if b.IncompleteKind() == cue.BottomKind {
		return a, nil
	}
	if a.IncompleteKind() == cue.BottomKind {
		return b, nil
	}

	r := &cue.Runtime{}

	// Build a new struct which will contain merged fields from A and B.
	def := &ast.StructLit{}

	// Iterate through A, grabbing fields from B.  We'll compare each field in
	// A to the fields in B, merging the values together.
	it, err := a.Fields(
		cue.All(),
		cue.Definitions(true),
		cue.Concrete(false),
	)
	if err != nil {
		return cue.Value{}, fmt.Errorf("error returning fields of a: %w", err)
	}

	// The general strategy is to:
	//
	// 1. Store a list of values - referring to type defintions - in A and B
	// 2. Dedupe the list
	// 3. Return a value referrnig to each type defintion left in the list.  If the
	//    list has a single value we can return the single value overall.  Else, we
	//    return a union of the values using a BinaryOp AST.
	//
	// There are two special cases:  structs and slices.
	//
	// If we're merging two single structs together, we recursively loop through each
	// field and merge the definitions together.  This is a special case.
	//
	// If we're merging slices together, we always create a binary op.

	for it.Next() {
		label := it.Label()

		aValue := it.Value()
		// Get the AST for the field in A, which is always an *ast.Field.
		aValAsField := aValue.Source().(*ast.Field)

		// Look the field up to check whether it's a member of the second struct B.
		// If it isn't found on the other struct, we can safely use A's field as-is.
		bField, _ := b.FieldByName(label, true)
		bValue := bField.Value

		if bValue.IncompleteKind() == cue.BottomKind {
			// Use A immediately, as there is no field in B.  Mark this field as
			// optional as it's only usable in one of the definitions.
			// TODO OPTIONAL
			def.Elts = append(def.Elts, aValAsField)
			continue
		}

		bValAsField := bValue.Source().(*ast.Field)

		// Get all values for the field, expanding the binary tree of unions into a
		// list of vales.
		aValues, err := expandValues(r, aValue)
		if err != nil {
			return cue.Value{}, err
		}
		bValues, err := expandValues(r, bValue)
		if err != nil {
			return cue.Value{}, err
		}

		// If we have one value each - and they're both structs - we need to recursively
		// merge these structs together into a single struct, containing optional fields
		// for values only represented in one.
		//
		// XXX (tonyhb): We could have a strategy which recursively merges _and_ returns
		// metadata about the similarities of structs.  We can then use this metadata to
		// determine whether to use a binary expression (eg. structs contain no overlap)
		// or to use the merged struct altogether.

		if len(aValues) <= 1 && len(bValues) <= 1 {
			// If the types are of different kinds we can immediately create a union
			// of the type definitions.
			if aValue.IncompleteKind() != bValue.IncompleteKind() {
				def.Elts = append(def.Elts, &ast.Field{
					Label: ast.NewIdent(label),
					Value: union(aValAsField.Value, bValAsField.Value),
				})
				continue
			}

			// If the values are of the same scalar kind, use the values from A.
			if scalarEquals(aValue.IncompleteKind(), bValue.IncompleteKind()) {
				// XXX (tonyhb): Are these concrete values?  If so, should we make an enum of
				// strings?
				//
				// The fields are the same scalar kind, so we can continue.
				def.Elts = append(def.Elts, aValAsField)
				continue
			}

			// If we're merging two structs together, recursively merge each member.
			if aValue.IncompleteKind() == cue.StructKind && bValue.IncompleteKind() == cue.StructKind {
				// Merge fields of the same complex kind recursively, eg. merge
				// two struct fields together into a new struct.
				next, err := recursivelyMerge(ctx, aValue, bValue)
				if err != nil {
					return cue.Value{}, err
				}

				switch src := next.Source().(type) {
				case *ast.Field:
					// We're returned an *ast.Field directly
					def.Elts = append(def.Elts, src)
					continue
				case *ast.File:
					// This is an *ast.File, as it's entirely formatted by cueutil.ASTToValue.
					// Pull put the declaration from the file.
					expr := next.Source().(*ast.File).Decls[0].(*ast.EmbedDecl).Expr
					def.Elts = append(def.Elts, &ast.Field{
						Label: ast.NewIdent(label),
						Value: expr,
					})
				default:
					return cue.Value{}, fmt.Errorf("unknown source kind for struct: %T", src)
				}

				// Continue on to the next field
				continue
			}

			// We're only left with merging slices.  This uses the same logic as merging >
			// 1 item from each value:  we're almost always goign to create a binary op
			// for the slice and deduplicate the op.
		}

		// Store a map of values we've seen.  The only way to do this nicely is
		// to format the AST into a well defined string.
		seen := map[string]struct{}{}
		deduped := []ast.Expr{}
		for _, item := range append(aValues, bValues...) {
			code, err := cueutil.ASTToSyntax(item.Source())
			if err != nil {
				return cue.Value{}, err
			}

			if _, ok := seen[code]; ok {
				continue
			}

			seen[code] = struct{}{}

			switch src := item.Source().(type) {
			case *ast.Field:
				deduped = append(deduped, src.Value)
			case *ast.File:
				deduped = append(deduped, src.Decls[0].(*ast.EmbedDecl).Expr)
			default:
				return cue.Value{}, fmt.Errorf("unknown ast type deduplicating value: %T", src)
			}
		}

		// Return a field containing all items.
		def.Elts = append(def.Elts, &ast.Field{
			Label: ast.NewIdent(label),
			Value: union(deduped...),
		})
		continue
	}

	// We may have missed fields from B that were missing in A.  Iterate through all of B's
	// fields and check which ones haven't been merged.
	it, err = b.Fields(
		cue.All(),
		cue.Definitions(true),
		cue.Concrete(false),
	)
	if err != nil {
		return cue.Value{}, fmt.Errorf("error returning fields for b: %w", err)
	}
	for it.Next() {
		label := it.Label()
		if a.LookupPath(cue.ParsePath(label)).IncompleteKind() != cue.BottomKind {
			// This field was found;  skip.
			continue
		}

		// This field isn't present in A, so it was skipped.  We can add this directly
		// to our struct.
		val := it.Value()
		aValAsField := val.Source().(*ast.Field)
		def.Elts = append(def.Elts, aValAsField)
	}

	return cueutil.ASTToValue(r, def)
}

// union merges all expressions into a single branched binary expression.
func union(elts ...ast.Expr) ast.Expr {
	if len(elts) == 1 {
		return elts[0]
	}

	current := &ast.BinaryExpr{
		Op: cue.OrOp.Token(),
	}
	for len(elts) > 0 {
		elt := elts[0]
		elts = elts[1:]

		current.X = elt
		if len(elts) == 1 {
			// The last item goes in the final branch.
			current.Y = elts[0]
			elts = elts[1:]
			continue
		}

		current.Y = &ast.BinaryExpr{
			Op: cue.OrOp.Token(),
		}
		current = current.Y.(*ast.BinaryExpr)
	}
	return current
}

func expandValues(r *cue.Runtime, union cue.Value) ([]cue.Value, error) {
	if union, ok := union.Syntax().(*ast.BinaryExpr); ok {
		vals := []cue.Value{}
		for _, expr := range expand(union) {
			val, err := cueutil.ASTToValue(r, expr)
			if err != nil {
				return vals, err
			}
			vals = append(vals, val)
		}
		return vals, nil
	}
	return []cue.Value{union}, nil
}

// expand walks a BinaryExpr, returning every non-binary expr as a single slice.
func expand(union *ast.BinaryExpr) []ast.Expr {
	result := []ast.Expr{}

	stack := []*ast.BinaryExpr{union}
	for len(stack) > 0 {
		item := stack[0]
		stack = stack[1:]

		if union, ok := item.X.(*ast.BinaryExpr); ok {
			stack = append(stack, union)
		} else {
			result = append(result, item.X)
		}

		if union, ok := item.Y.(*ast.BinaryExpr); ok {
			stack = append(stack, union)
		} else {
			result = append(result, item.Y)
		}
	}

	return result
}

func isScalar(k cue.Kind) bool {
	return k != cue.StructKind && k != cue.ListKind
}

func scalarEquals(a, b cue.Kind) bool {
	if !isScalar(a) || !isScalar(b) {
		// Structs and arrays can never be the same by shallowc comparisons;  they may
		// have different members.
		return false
	}
	return a == b
}

func equals(a, b cue.Value) bool {
	// If A or B is a slice, we can't use subsumes as it's buggy:  a slice of one type
	// subsumes a slice of another.
	//
	// See https://github.com/cue-lang/cue/issues/1654 for more info.
	return (a.Subsume(b, cue.All()) == nil || b.Subsume(a, cue.All()) == nil) && (a.Unify(b).Kind() != cue.BottomKind)
}
