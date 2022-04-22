package fromjson

import (
	"fmt"
	"math"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/token"
	"github.com/inngest/event-schemas/pkg/cueutil"
)

// FromJSON takes a JSON map and generates a CUE type which validates the given
// values.  Because this works backwards from values, it will never generate
// constraints and will likely contain errors with eg. null values. However,
// the cue type that it generates is a good starting point for customizing
// and creating a properly defined type for the value.
//
// This produces a struct with no top-level identifier:
//
// {
//    name: string
// }
//
// NOTE: Because maps are unordered in Go, the resulting type will have fields
// returned in no specific order.  The type will, however, match the iteration
// order from Go.
func FromJSON(input map[string]interface{}) (typeDef string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error generating type from JSON: %v", r)
		}
	}()

	// Create a new struct that contains this type.
	def := ast.NewStruct()

	// Walk the type recursively, adding fields to the type definition/.
	walk(input, def)

	// Format the cue code.
	byt, err := format.Node(
		def,
		format.TabIndent(false),
		format.UseSpaces(2),
	)
	if err != nil {
		return "", fmt.Errorf("error formatting generated type: %w", err)
	}

	return string(byt), err
}

// walk iterates through the given JSON object, creating a new cue type definition for
// each key within the map.
//
// It does this by constructing the correct AST for each field within the struct, then
// generating AST representing the field's type.  We do not generate positions for any
// of the AST's syntax;  instead, we rely on cue's format package to autogen the right
// syntax for us.
//
// NOTE: We short-circuit some more complex AST for enums when generating the AST for
// each type.  See typeExpr for more information.
//
// If this needs to be modified, a starting point for inspecting cue's AST is:
//
//	r := &cue.Runtime{}
//	// compile some cue types here.
//	i, err := r.Compile(".", `{ cue: string, type: int, here: bool }`)
//	if err != nil {
//		return err
//	}
//	node := i.Value().Source()
//	// inspect the AST.
//	spew.Dump(node)
func walk(obj map[string]interface{}, def *ast.StructLit) {

	for k, v := range obj {
		typ := kind(v)

		// Generate a field for this key in the struct.
		var value ast.Expr
		switch typ {
		case cue.ListKind:
			var structs []*ast.StructLit

			typ, structs = walkSlice(v.([]interface{}))
			if typ == cue.StructKind && len(structs) == 1 {
				// This is an array of structs.
				value = &ast.ListLit{
					Elts: []ast.Expr{
						&ast.Ellipsis{
							Type: structs[0],
						},
					},
				}

			}
			if typ == cue.StructKind && len(structs) > 1 {
				// There is more than one struct.  Make binary expressions
				// for all structs.
				current := &ast.BinaryExpr{
					Op: cue.OrOp.Token(),
				}
				top := current
				for n, s := range structs {
					copied := s
					if n == len(structs)-1 {
						// This is the last element, so mark it as the Y of the final
						// binary expression.
						current.Y = copied
						continue
					}
					current.X = copied
					if n < len(structs)-2 {
						// There is > 1 item left, so we need a new binary
						// expression to join the next two.
						current.Y = &ast.BinaryExpr{
							Op: cue.OrOp.Token(),
						}
						current = current.Y.(*ast.BinaryExpr)
					}
				}

				value = &ast.ListLit{
					Elts: []ast.Expr{
						&ast.Ellipsis{
							Type: top,
						},
					},
				}
			}

			if typ != cue.StructKind || len(structs) == 0 {
				value = &ast.ListLit{
					Elts: []ast.Expr{
						&ast.Ellipsis{
							Type: typeExpr(typ),
						},
					},
				}
			}
		case cue.StructKind:
			// Create a new struct and walk the map
			inner := ast.NewStruct()
			walk(v.(map[string]interface{}), inner)
			value = inner
		default:
			// by default this is a basic type, eg "string".  Use
			// the type generated from the value as the field's type.
			value = typeExpr(typ)
		}

		// Add this field to the cue struct.
		def.Elts = append(def.Elts, &ast.Field{
			Label: ast.NewIdent(k),
			Value: value,
		})
	}
}

// typeExpr returns an ast.Expr representing the type for the given cue kind.
func typeExpr(k cue.Kind) ast.Expr {
	// Usually, Cue's syntax here is an ast.NewIdent.  However, typ.TypeString()
	// returns "(string | int)" for compound types, and to represent this we'd need
	// to construct a compund *ast.Ident containing binary operators.  EG, [...string | int | float]
	// is actually:
	//
	//
	// &ast.Ellipses{Type: &ast.BinaryExpr{X: &ast.BinaryExpr{X: ast.NewIdent("string"), ...}}}
	//
	// Note the nested binary expressions representing each enum.
	//
	// Because we're dumping a string version of the syntax, we can straight up use the string here:
	// it returns the same string output as the more complex type.
	return &ast.BasicLit{Kind: token.STRING, Value: k.TypeString()}
}

// walkSlice walks a slice to calculate the types which occur within the slice.  If the
// slice is primitive scalars only the types are represented by the single cue.Kind type
// returned (via a bitmask).  If the slice contains a struct, we return the cue.Kind
// bitmask and a list of []*ast.StructLit struct definitons.
func walkSlice(slice []interface{}) (cue.Kind, []*ast.StructLit) {
	var found cue.Kind

	structs := []*ast.StructLit{}
	for _, item := range slice {
		k := kind(item)

		if k == cue.StructKind {
			// Map the type of this struct.
			structAST := ast.NewStruct()
			walk(item.(map[string]interface{}), structAST)
			structs = append(structs, structAST)
		}

		if found == cue.BottomKind {
			found = k
			continue
		}

		// Add the type to the list of available types.
		found = found | k
	}

	if found == cue.BottomKind {
		// no type;  return "_" for any.
		return cue.TopKind, nil
	}

	// Deduplicate struct definitions by seeing which are subsumable.
	// We can't rely on ASTs as maps have randomized key ordering.
	r := &cue.Runtime{}

	deduped := []*ast.StructLit{}
NEXT:
	for _, next := range structs {
		if len(deduped) == 0 {
			deduped = append(deduped, next)
		}

		// We ignore errors as this is best-effort.  Worst case we return
		// no concrete struct definitions and use the top-level {...}
		// struct identifier for any key/values.
		instA, _ := cueutil.ASTToValue(r, next)

		// Does this match any existing struct type?
		for _, existing := range deduped {
			// XXX: Store these mapped to process once.
			instB, _ := astToValue(r, existing)

			subA := instA.Value().Subsumes(instB.Value())
			subB := instB.Value().Subsumes(instA.Value())
			if subA && subB {
				// This is the same as an existing type.  Continue
				// the iteration through struct definitions.
				continue NEXT
			}
		}

		// This doesn't match any, so we add and continue
		deduped = append(deduped, next)
	}

	return found, deduped
}

// kind returns a cue.Kind representing the type for the given value.
func kind(v interface{}) cue.Kind {
	switch cast := v.(type) {
	case float64:
		// All JSON numbers are represented as float64, even if they're ints.
		// We want to accurately check whether this is a float or an int.
		_, frac := math.Modf(cast)
		if frac == 0.0 {
			return cue.IntKind
		}
		return cue.FloatKind
	case bool:
		return cue.BoolKind
	case string:
		return cue.StringKind
	case map[string]interface{}:
		return cue.StructKind
	case []interface{}:
		return cue.ListKind
	case interface{}:
		return cue.TopKind
	case nil:
		// Use the TopKind here as well as null, as null may be another value in the future.
		//
		// This automatically gets represented as "_", but at least we can track
		// internally the type.
		return cue.NullKind | cue.TopKind
	}

	return cue.TopKind
}
