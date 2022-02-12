package typescript

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
)

var (
	ctxIndentLevel = "indent"
)

func MarshalString(cuestr string) (string, error) {
	r := &cue.Runtime{}
	inst, err := r.Compile(".", cuestr)
	if err != nil {
		return "", fmt.Errorf("error generating inst: %w", err)
	}

	return MarshalCueValue(inst.Value())

}

// MarshalCueValue returns a typescript type given a cue value.
func MarshalCueValue(v cue.Value) (string, error) {
	if err := v.Validate(); err != nil {
		return "", err
	}

	// Assume that this is a top-level object containing all definitions,
	// and iterate over each definition
	it, err := v.Fields(cue.Definitions(true), cue.Concrete(false))
	if err != nil {
		return "", err
	}

	exprs := []*Expr{}

	for it.Next() {
		result, err := generateExprs(context.Background(), it.Label(), it.Value())
		if err != nil {
			return "", err
		}
		exprs = append(exprs, result...)
	}

	str, err := FormatAST(exprs...)
	fmt.Println("---")
	fmt.Println(str)
	return str, err
}

// generateExprs creates a typescript expression for a top-level identifier.  This
// differs to the 'generateAST' function as it wraps the created AST within an Expr,
// representing a complete expression terminating with a semicolon.
func generateExprs(ctx context.Context, label string, v cue.Value) ([]*Expr, error) {
	ast, err := generateAST(ctx, label, v)
	if err != nil {
		return nil, err
	}

	exprs := make([]*Expr, len(ast))
	for n, a := range ast {
		// Wrap the AST in an expression, indicating that the AST is a
		// fully defined typescript expression.
		if _, ok := a.(Local); ok {
			exprs[n] = &Expr{Data: a}
			continue
		}

		kind := LocalType
		if binding, ok := a.(Binding); ok && binding.Kind == BindingType {
			// If we're making a struct type, use an Interface declaration
			// instead of the default `const Event = type {` declaration.
			kind = LocalInterface
		}

		exprs[n] = &Expr{
			Data: Local{
				Kind:     kind,
				Name:     label,
				IsExport: true,
				Value:    a,
			},
		}
	}

	return exprs, nil
}

// generateAST creates typescript AST for the given cue values.
func generateAST(ctx context.Context, label string, v cue.Value) ([]AstKind, error) {
	// We have the cue's value, although this may represent many things.
	// Notably, v.IncompleteKind() returns cue.StringKind even if this field
	// represents a static string, a string type, or an enum of strings.
	//
	// In order to properly generate Typescript AST for the value we need to
	// walk Cue's AST.
	fmt.Printf("%s: %T\n", label, v.Syntax())

	switch v.IncompleteKind() {
	case cue.StructKind:
		ast, err := generateStructBinding(ctx, v)
		if err != nil {
			return nil, err
		}
		return ast, nil
	case cue.ListKind:
		// TODO: HANDLE LISTS
		return nil, nil
	default:
		syn := v.Syntax(cue.All())
		switch ident := syn.(type) {
		case *ast.BinaryExpr:
			// This could be an enum, a basic lit with a constraint, or a type Ident
			// with a constraint.
			//
			// XXX: Add constraints as comments above the identifier.
			// XXX: We could also generate functions which validate constraints
			// with an aexpression
			return generateEnum(ctx, label, v)
		case *ast.BasicLit:
			// This is a const.
			scalar, err := generateScalar(ctx, label, v)
			if err != nil {
				return nil, err
			}
			return []AstKind{scalar}, nil
		case *ast.Ident:
			return []AstKind{
				Type{Value: identToTS(ident.Name)},
			}, nil
		default:
			return nil, fmt.Errorf("unhandled cue type: %T", ident)
		}
	}
}

// generateLocal returns a scalar identifier, such as a top-level const
// or top-level type.
func generateScalar(ctx context.Context, label string, v cue.Value) (AstKind, error) {
	var i interface{}
	if err := v.Decode(&i); err != nil {
		return nil, err
	}
	return Scalar{Value: i}, nil
}

func generateEnum(ctx context.Context, label string, v cue.Value) ([]AstKind, error) {
	// We know that v is of type Expr, and it contains a BinaryExpr.
	_, _ = v.Expr()
	fmt.Println("ENUM", v)
	return nil, nil
}

// generateStructBinding returns a binding representing a TypeScript object
// or interface.
//
// It does not wrap this within a Local as this function is used within top-level
// and nested structs;  nested structs are the Value of a KeyValue whereas
// top-level identifiers are values of a Local.
func generateStructBinding(ctx context.Context, v cue.Value) ([]AstKind, error) {
	it, err := v.Fields(cue.All())
	if err != nil {
		return nil, err
	}

	// Referencing another field.
	// p, r := v.Reference()
	//
	// A struct may contain enum definions.  Within typescript we want to pull
	// those enum definitions to top-level types.  This means that we must return
	// multiple top-level AST expressions.

	ast := []AstKind{}
	for it.Next() {
		if it.IsHidden() {
			continue
		}

		// Create the raw AST for each field's value
		created, err := generateAST(withIncreasedIndentLevel(ctx), it.Label(), it.Value())
		if err != nil {
			return nil, err
		}

		// Wrap the AST value within a KeyValue.
		wrapped := make([]AstKind, len(created))
		for n, item := range created {
			wrapped[n] = KeyValue{
				Key:      it.Label(),
				Value:    item,
				Optional: it.IsOptional(),
			}
		}

		ast = append(ast, wrapped...)
	}

	return []AstKind{Binding{
		Kind:        BindingType,
		Members:     ast,
		IndentLevel: indentLevel(ctx),
	}}, nil
}

func identToTS(name string) string {
	switch name {
	case "bool":
		return "boolean"
	case "float", "int", "number":
		return "number"
	case "_":
		return "unknown"
	case "[...]":
		return "Array<unknown>"
	case "{...}":
		return "{ [key: string]: unknown }"
	default:
		return name
	}
}

// indentLevel returns the current indent level from the context.  This is
// a quick and dirty way of formatting nested structs.
func indentLevel(ctx context.Context) int {
	indent, _ := ctx.Value(ctxIndentLevel).(int)
	return indent
}

func withIncreasedIndentLevel(ctx context.Context) context.Context {
	level := indentLevel(ctx) + 1
	return context.WithValue(ctx, ctxIndentLevel, level)
}
