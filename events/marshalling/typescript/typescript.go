package typescript

import (
	"context"
	"fmt"
	"strings"

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
		if len(exprs) > 0 {
			// Add two newlines between each field.
			exprs = append(exprs, []*Expr{
				{Data: Lit{Value: "\n"}},
				{Data: Lit{Value: "\n"}},
			}...)
		}

		result, err := generateExprs(context.Background(), it.Label(), it.Value())
		if err != nil {
			return "", err
		}
		exprs = append(exprs, result...)
	}

	// Add a final newline to terminate the file.
	exprs = append(exprs, []*Expr{{Data: Lit{Value: "\n"}}}...)

	str, err := FormatAST(exprs...)
	return str, err
}

// generateExprs creates a typescript expression for a top-level identifier.  This
// differs to the 'generateAST' function as it wraps the created AST within an Expr,
// representing a complete expression terminating with a semicolon.
func generateExprs(ctx context.Context, label string, v cue.Value) ([]*Expr, error) {
	label = strings.Title(strings.ToLower(label))

	exprs, ast, err := generateAST(ctx, label, v)
	if err != nil {
		return nil, err
	}

	for _, a := range ast {
		// Wrap the AST in an expression, indicating that the AST is a
		// fully defined typescript expression.
		if _, ok := a.(Local); ok {
			exprs = append(exprs, &Expr{Data: a})
			continue
		}

		if enum, ok := a.(Enum); ok {
			// Enums define their own top-level Local AST as they create
			// more than one export.
			enumExprs, err := enum.AST()
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, enumExprs...)
			continue
		}

		kind := LocalType
		if binding, ok := a.(Binding); ok && binding.Kind == BindingType {
			// If we're making a struct type, use an Interface declaration
			// instead of the default `const Event = type {` declaration.
			kind = LocalInterface
		}

		exprs = append(exprs, &Expr{
			Data: Local{
				Kind:     kind,
				Name:     label,
				IsExport: true,
				Value:    a,
			},
		})
	}

	return exprs, nil
}

// generateAST creates typescript AST for the given cue values
//
// If the value contains a field of enums, this may generate top-level expressions
// to add to the generated typescript file.
func generateAST(ctx context.Context, label string, v cue.Value) ([]*Expr, []AstKind, error) {
	// We have the cue's value, although this may represent many things.
	// Notably, v.IncompleteKind() returns cue.StringKind even if this field
	// represents a static string, a string type, or an enum of strings.
	//
	// In order to properly generate Typescript AST for the value we need to
	// walk Cue's AST.
	// fmt.Printf("%s: %T\n", label, v.Syntax())

	switch v.IncompleteKind() {
	case cue.StructKind:
		exprs, ast, err := generateStructBinding(ctx, v)
		if err != nil {
			return nil, nil, err
		}
		return exprs, ast, nil
	case cue.ListKind:
		// TODO: HANDLE LISTS
		return nil, nil, nil
	default:
		syn := v.Syntax(cue.All())
		switch ident := syn.(type) {
		case *ast.BinaryExpr:
			// This could be an enum, a basic lit with a constraint, or a type Ident
			// with a constraint.
			op, _ := v.Expr()
			if op == cue.OrOp {
				// This is an enum.  We're combining > 1 field using
				// the Or operator.
				ast, err := generateEnum(ctx, label, v)
				return nil, ast, err
			}

			if op == cue.AndOp {
				// Although it's possible to combine two structs via the AndOp,
				// those are handled within the IncompleteKind() check above.
				//
				// Because of this we can guarantee that this is a constrained
				// type check.
				ast, err := generateConstrainedIdent(ctx, label, v)
				return nil, ast, err
			}
		case *ast.BasicLit:
			// This is a const.
			scalar, err := generateScalar(ctx, label, v)
			if err != nil {
				return nil, nil, err
			}
			return nil, []AstKind{scalar}, nil
		case *ast.Ident:
			return nil, []AstKind{
				Type{Value: identToTS(ident.Name)},
			}, nil
		default:
			return nil, nil, fmt.Errorf("unhandled cue ident: %T", ident)
		}
	}
	return nil, nil, fmt.Errorf("unhandled cue type: %v", v.IncompleteKind())
}

func generateConstrainedIdent(ctx context.Context, label string, v cue.Value) ([]AstKind, error) {
	// All types being constrained should share the same heirarchy - eg. uint refers to an
	// int and a number.  In Typescript we don't really care about refined types and can
	// use the first value, as typescript only uses "number" or "string" etc..
	//
	// XXX: Add constraints as comments above the identifier. We could also generate
	// functions which validate constraints with an aexpression
	_, vals := v.Expr()
	_, ast, err := generateAST(ctx, label, vals[0])
	return ast, err
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
	label = strings.Title(strings.ToLower(label))

	_, vals := v.Expr()
	members := make([]AstKind, len(vals))

	// Generate AST representing the value of each member in the enum.
	for n, val := range vals {
		_, ast, err := generateAST(ctx, label, val)
		if err != nil {
			return nil, fmt.Errorf("error generating ast for enum val: %w", err)
		}
		if len(ast) > 1 {
			return nil, fmt.Errorf("invalid ast generated for enum val: %v", ast)
		}
		members[n] = ast[0]
	}

	return []AstKind{Enum{
		Name:    label,
		Members: members,
	}}, nil
}

// generateStructBinding returns a binding representing a TypeScript object
// or interface.
//
// It does not wrap this within a Local as this function is used within top-level
// and nested structs;  nested structs are the Value of a KeyValue whereas
// top-level identifiers are values of a Local.
func generateStructBinding(ctx context.Context, v cue.Value) ([]*Expr, []AstKind, error) {
	it, err := v.Fields(cue.All())
	if err != nil {
		return nil, nil, err
	}

	expr := []*Expr{}

	members := []AstKind{}
	for it.Next() {
		if it.IsHidden() {
			continue
		}

		// Create the raw AST for each field's value
		newExpr, created, err := generateAST(withIncreasedIndentLevel(ctx), it.Label(), it.Value())
		expr = append(expr, newExpr...)
		if err != nil {
			return nil, nil, err
		}

		if len(created) == 0 {
			continue
		}

		// We may have generated top-level local definitions, which we should pull out
		// to the AST context and not use as a key-value.
		if local, ok := created[0].(Local); ok {
			// Add the fields to the top-level object being created.
			expr = append(expr, &Expr{Data: created[0]})
			// And add a reference to the type as the key value pair.
			created[0] = Type{Value: local.Name}
		}

		// A struct may contain enum definions.  Within typescript we want to pull
		// those enum definitions to top-level types.  This means that we must return
		// multiple top-level AST expressions.
		if enum, ok := created[0].(Enum); ok {
			enumAst, err := enum.AST()
			if err != nil {
				return nil, nil, err
			}
			expr = append(expr, enumAst...)
			// Add two newlines between each enum and struct visually.
			expr = append(expr, []*Expr{
				{Data: Lit{Value: "\n"}},
				{Data: Lit{Value: "\n"}},
			}...)
			// Use the enum name as the key's value.
			created[0] = Type{Value: enum.Name}
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

		members = append(members, wrapped...)
	}

	// Add the struct to the generated AST fields.
	ast := []AstKind{Binding{
		Kind:        BindingType,
		Members:     members,
		IndentLevel: indentLevel(ctx),
	}}

	return expr, ast, nil
}

// identToTS returns Typescript type names from a given cue type name.
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

// withIncreasedIndentLevel increases the indent level in the given context,
// returning a new context with the updated indent level.
func withIncreasedIndentLevel(ctx context.Context) context.Context {
	level := indentLevel(ctx) + 1
	return context.WithValue(ctx, ctxIndentLevel, level)
}
