package fromjson

import (
	"fmt"
	"math"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/token"
)

// FromJSON takes a JSON map and generates a CUE type which validates the given
// values.  Because this works backwards from values, it will never generate
// constraints and will likely contain errors with eg. null values.
//
// However, the cue type that it generates is a good starting point for customizing
// and creating a properly defined type for the value.
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
			typ = walkSlice(v.([]interface{}))
			value = &ast.ListLit{
				Elts: []ast.Expr{
					&ast.Ellipsis{
						Type: typeExpr(typ),
					},
				},
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

// walkSlice walks a slice to calculate the types which occur within the slice.
func walkSlice(slice []interface{}) cue.Kind {
	var found cue.Kind

	for _, item := range slice {
		if found == cue.BottomKind {
			found = kind(item)
			continue
		}
		// Add the type to the list of available types.
		found = found | kind(item)
	}

	if found == cue.BottomKind {
		// no type;  return "_" for any.
		return cue.TopKind
	}

	return found
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
