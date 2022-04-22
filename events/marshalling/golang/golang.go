package golang

import (
	"context"
	"fmt"
	"go/ast"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"github.com/inngest/event-schemas/events/marshalling"
)

func MarshalString(cuestr string) (string, error) {
	r := &cue.Runtime{}
	inst, err := r.Compile(".", cuestr)
	if err != nil {
		return "", fmt.Errorf("error generating inst: %w", err)
	}

	return MarshalCueValue(inst.Value())
}

func MarshalCueValue(v cue.Value) (string, error) {
	exprs, err := marshalling.Walk(v, handleDefinition)
	if err != nil {
		return "", nil
	}
	return marshalling.Format(exprs...)
}

// handleDefinition creates a golang type for a top-level identifier
func handleDefinition(ctx context.Context, label string, v cue.Value) ([]marshalling.Expr, error) {
	switch v.IncompleteKind() {
	case cue.StructKind:
		exprs, ast, err := generateStructBinding(ctx, v)
		if err != nil {
			return nil, nil, err
		}
		return exprs, ast, nil
	case cue.ListKind:
		return generateArray(ctx, label, v)
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

// generateAST creates typescript AST for the given cue values
//
// If the value contains a field of enums, this may generate top-level expressions
// to add to the generated typescript file.
func generateAST(ctx context.Context, label string, v cue.Value) ([]marshalling.Expr, []AstKind, error) {
	// We have the cue's value, although this may represent many things.
	// Notably, v.IncompleteKind() returns cue.StringKind even if this field
	// represents a static string, a string type, or an enum of strings.
	//
	// In order to properly generate Typescript AST for the value we need to
	// walk Cue's AST.
	switch v.IncompleteKind() {
	case cue.StructKind:
		exprs, ast, err := generateStructBinding(ctx, v)
		if err != nil {
			return nil, nil, err
		}
		return exprs, ast, nil
	case cue.ListKind:
		return generateArray(ctx, label, v)
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
