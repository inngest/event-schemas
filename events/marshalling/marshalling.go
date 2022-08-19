package marshalling

import (
	"context"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
)

type Generator interface {
	// AST is the entrypoint when parsing any AST.  This delegates
	// to the specific functions below.
	//
	// It is given the context, the current AST to generate an expr from,
	// plus all parsed AST items and all currently generated Exprs
	AST(context.Context, []ParsedAST) ([]Expr, error)
}

func Marshal(ctx context.Context, v cue.Value, g Generator) (string, error) {
	parsed, err := Parse(ctx, v)
	if err != nil {
		return "", err
	}

	exprs, err := g.AST(ctx, parsed)
	if err != nil {
		return "", err
	}

	return Format(exprs...)
}

// Parse walks through a cue value, typically from a cue.Instance, calling the
// parsing the type definitions into our interim AST.  The interim AST makes
// it much easier to generate basic types in many languages.
func Parse(ctx context.Context, v cue.Value) ([]ParsedAST, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}

	// Assume that this is a top-level object containing all definitions,
	// and iterate over each definition
	it, err := v.Fields(cue.Definitions(true), cue.Concrete(false))
	if err != nil {
		return nil, err
	}

	// This is a two-pass method.  First, we iterate through each value within
	// the cue definition to create a parsed AST.
	//
	// This is "kind of" a tree;  each parsed AST item can contain nested members.
	//
	// We then call generate on the parsed AST.

	parsed := []ParsedAST{}
	for it.Next() {
		// Here we want to generalize the type definitions from cue into interim
		// ASTs that are easier to modify.
		//
		// Each language we want to implement requires doing the same thing:  inspecting
		// the cue AST, reading it correctly to parse eg. enums from binary ops, then
		// creating the correct types from the parsed ASTs.
		//
		// Instead of doing this per-language, we handle this once here.

		p, err := parseAST(ctx, it.Label(), it.Value())
		if err != nil {
			return parsed, err
		}
		parsed = append(parsed, p)
	}

	// Now that we have our parsed AST, we send this into the generator
	// to generate Exprs, which are raw structs representing concrete type
	// syntax to be formatted.
	return parsed, err
}

func parseAST(ctx context.Context, label string, v cue.Value) (ParsedAST, error) {
	// We have the cue's value, although this may represent many things.
	// Notably, v.IncompleteKind() returns cue.StringKind even if this field
	// represents a static string, a string type, or an enum of strings.
	//
	// In order to properly generate Typescript AST for the value we need to
	// walk Cue's AST.
	switch v.IncompleteKind() {
	case cue.StructKind:
		s, err := parseStruct(ctx, v)
		if err != nil {
			return nil, err
		}
		s.name = label
		return s, nil
	case cue.ListKind:
		arr, err := parseArray(ctx, label, v)
		if err != nil {
			return nil, err
		}
		arr.name = label
		return arr, nil
	default:
		// We have to iterate through the actual syntax / AST of the
		// cue value in order to determine other type.
		//
		// This is separated into a separate function as its re-used when
		// creating arrays.
		return parseCueSyntax(ctx, label, v, v.Syntax(cue.All()))
	}
}

func parseCueSyntax(ctx context.Context, label string, v cue.Value, syn ast.Node) (ParsedAST, error) {
	switch ident := syn.(type) {
	case *ast.UnaryExpr:
		// This is a default value.
		return nil, nil
	case *ast.BinaryExpr:
		// This could be an enum, a basic lit with a constraint, or a type Ident
		// with a constraint, or something with a default value.
		op, _ := v.Expr()
		switch op {
		case cue.OrOp:
			// This is an enum.  We're combining > 1 field using
			// the Or operator.
			enum, err := parseEnum(ctx, label, v)
			if err != nil {
				return nil, err
			}
			enum.name = label
			return enum, nil
		case cue.AndOp:
			// Although it's possible to combine two structs via the AndOp,
			// those are handled within the IncompleteKind() check above.
			//
			// Because of this we can guarantee that this is a constrained
			// type check.
			return parseConstraintedIdent(ctx, label, v)

		case cue.NoOp:
			// A leaf in the binary expression could hold the default value.
			// var defValue ParsedAST
			enum := &ParsedEnum{
				name:    label,
				Members: []ParsedAST{},
			}
			nodes := []ast.Expr{ident.X, ident.Y}

			// Y could always be a unary expression, which is a default
			// default value.  This needs to be special-cased.
			if uexp, ok := ident.Y.(*ast.UnaryExpr); ok {
				val, err := astToValue(ident.X)
				if err != nil {
					return nil, err
				}
				def, err := astToValue(uexp.X)
				if err != nil {
					return nil, err
				}
				parsed, err := parseAST(ctx, label, val)
				if err != nil {
					return nil, err
				}
				parsedDefault, err := parseAST(ctx, "", def)
				if err != nil {
					return nil, err
				}
				parsed.SetDefault(parsedDefault)
				return parsed, nil
			}

			for len(nodes) > 0 {
				node := nodes[0]
				nodes = nodes[1:]
				// If it's a binary expr, add all nodes to the queue.
				if bexp, ok := node.(*ast.BinaryExpr); ok {
					// This odd appending is to retain the order of the
					// array as we parse the binarbinary expressions.
					nodes = append(
						[]ast.Expr{bexp.X},
						append([]ast.Expr{bexp.Y}, nodes...)...,
					)
					continue
				}

				v, err := astToValue(node)
				if err != nil {
					return nil, fmt.Errorf("error parsing enum ast: %w", err)
				}
				parsed, err := parseAST(ctx, "", v)
				if err != nil {
					return nil, err
				}
				enum.Members = append(enum.Members, parsed)
			}
			return enum, nil
		}

		// Otherwise... what are we doing here?
		return nil, fmt.Errorf("unknown binary op: %#v", op)
	case *ast.BasicLit:
		// This is a concrete value.
		// Convert this to a value for decoding into the concrete type.
		value, err := astToValue(ident)
		if err != nil {
			return nil, fmt.Errorf("error converting syntax struct to value: %w", err)
		}
		return parseScalar(ctx, label, value)
	case *ast.Ident:
		return &ParsedIdent{
			name:  label,
			Ident: ident,
		}, nil
	case *ast.StructLit:
		// Convert this to a value then iterate through the cue.StructKind value.
		value, err := astToValue(ident)
		if err != nil {
			return nil, fmt.Errorf("error converting syntax struct to value: %w", err)
		}
		return parseStruct(ctx, value)
	default:
		return nil, fmt.Errorf("unhandled cue type: %v (%T)", v.IncompleteKind(), ident)
	}
}

// parseStruct returns a ParsedStruct with all of the cue fields parsed as
// members.
func parseStruct(ctx context.Context, v cue.Value) (*ParsedStruct, error) {
	parsed := &ParsedStruct{
		Members: []*ParsedStructField{},
	}

	it, err := v.Fields(cue.All())
	if err != nil {
		return parsed, err
	}

	for it.Next() {
		if it.IsHidden() {
			// We do not currently export this.
			//
			// XXX: Should this be included as an unexported field,
			// eg within golang we can create a non-exported lowercase field name?
			continue
		}

		// Create the raw AST for each field's value
		member, err := parseAST(ctx, it.Label(), it.Value())
		if err == nil && member == nil {
			// XXX: When would we create a nil member?
			continue
		}

		if err != nil {
			return parsed, err
		}

		parsed.Members = append(parsed.Members, &ParsedStructField{
			ParsedAST: member,
			Optional:  it.IsOptional(),
		})
	}

	return parsed, nil
}

// parseArray returns an array.  This will always produce a type definition, even if all
// values in the cue list are basic literal values (eg. instead of ["1", "2"] this will generate
// Array<string>).
//
// This may return top-level expressions if the array contains a struct with enums.
func parseArray(ctx context.Context, label string, v cue.Value) (*ParsedArray, error) {
	var err error
	parsed := &ParsedArray{}

	// If this is a binary array, this is a list with a default.
	bexpr, ok := v.Syntax(cue.All()).(*ast.BinaryExpr)
	if ok {
		// Y stores the UnaryExpr default, and X is the array.
		v, err = astToValue(bexpr.X)
		if err != nil {
			return nil, err
		}
		if uexpr, ok := bexpr.Y.(*ast.UnaryExpr); ok {
			defVal, err := astToValue(uexpr.X)
			if err != nil {
				return nil, err
			}
			parsed.Default, err = parseAST(ctx, "", defVal)
			if err != nil {
				return nil, fmt.Errorf("error parsing default: %w", err)
			}
		}
	}

	// We can't use v.List() to create an iterator as iterators don't return types:
	// they only work with concrete values (which give us concrete scalars).
	//
	// Instead, take the Cue AST and walk it to determine the elements in the list,
	// and create TS AST from them.
	listLit, ok := v.Syntax(cue.All()).(*ast.ListLit)
	if !ok {
		return parsed, fmt.Errorf("unknown list ast type: %T", v.Syntax(cue.All()))
	}

	if len(listLit.Elts) == 0 {
		// We're done.
		return parsed, nil
	}

	elts := listLit.Elts
	if ellipsis, ok := listLit.Elts[0].(*ast.Ellipsis); ok {
		// TODO: ???
		elts = []ast.Expr{ellipsis.Type}
	}

	for len(elts) > 0 {
		elt := elts[0]
		elts = elts[1:]

		p, err := parseCueSyntax(ctx, "", v, elt)
		if err != nil {
			return parsed, err
		}
		parsed.Members = append(parsed.Members, p)
	}
	return parsed, nil
}

// parseEnum creates an enum definition which should be epanded to its
// full Expr AST for a given value.
func parseEnum(ctx context.Context, label string, v cue.Value) (*ParsedEnum, error) {
	enum := &ParsedEnum{
		name:    label,
		Members: []ParsedAST{},
	}

	_, vals := v.Expr()
	// Generate AST representing the value of each member in the enum.
	for _, val := range vals {
		ast, err := parseAST(ctx, "", val)
		if err != nil {
			return enum, fmt.Errorf("error generating ast for enum val: %w", err)
		}
		enum.Members = append(enum.Members, ast)
	}

	return enum, nil
}

// parseScalar returns a parsed scalar, such as a top-level const
// or top-level type.
func parseScalar(ctx context.Context, label string, v cue.Value) (ParsedAST, error) {
	var i interface{}
	if err := v.Decode(&i); err != nil {
		return nil, err
	}
	return &ParsedScalar{name: label, Value: i}, nil
}

func parseConstraintedIdent(ctx context.Context, label string, v cue.Value) (ParsedAST, error) {
	// TODO: Parse constraints and include within the ident.
	_, vals := v.Expr()
	// Hack: take the first value which is always the constrained ident.
	return parseAST(ctx, label, vals[0])
}

func Format(expr ...Expr) (string, error) {
	str := strings.Builder{}

	for _, e := range expr {
		if _, err := str.WriteString(e.String()); err != nil {
			return "", err
		}
	}

	return str.String(), nil
}

// Expr represents an expression in any language, used to construct
// the generated type.
type Expr interface {
	fmt.Stringer
}

// Lit represents a literal string value, which is returned as-is.
type Lit struct {
	Value string
}

// String fulfils the Expr interface, returning the string value.
func (l Lit) String() string { return l.Value }

func astToValue(ast ast.Node) (cue.Value, error) {
	r := &cue.Runtime{}
	// XXX: We really need a better way to create a cue.Value from
	// an AST struct.
	byt, _ := format.Node(
		ast,
		format.TabIndent(false),
		format.UseSpaces(2),
	)
	inst, err := r.Compile(".", byt)
	if err != nil {
		return cue.Value{}, err
	}
	return inst.Value(), nil
}
