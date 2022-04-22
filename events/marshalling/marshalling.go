package marshalling

import (
	"context"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
)

type Generator interface {
	// Enum handles the generation of an enum value.
	Enum(ctx context.Context, label string, v cue.Value)
	Struct(ctx context.Context, label string, v cue.Value)
	ConstrainedIdent(ctx context.Context, label string, v cue.Value)
	Ident(ctx context.Context, label string, v cue.Value)
	Scalar(ctx context.Context, label string, v cue.Value)
}

type WalkerFunc func(ctx context.Context, field string, val cue.Value) ([]Expr, error)

// Walk walks through a cue vaalue, typically from a cue.Instance, calling the
// supplied WalkerFunc on each field (definition) within the cue value.
func Walk(v cue.Value, w WalkerFunc) ([]Expr, error) {
	ctx := context.Background()

	if err := v.Validate(); err != nil {
		return nil, err
	}

	// Assume that this is a top-level object containing all definitions,
	// and iterate over each definition
	it, err := v.Fields(cue.Definitions(true), cue.Concrete(false))
	if err != nil {
		return nil, err
	}

	exprs := []Expr{}

	for it.Next() {
		if len(exprs) > 0 {
			// Add two newlines between each field.
			exprs = append(exprs, []Expr{
				Lit{Value: "\n"},
				Lit{Value: "\n"},
			}...)
		}

		result, err := w(ctx, it.Label(), it.Value())
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, result...)
	}

	// Add a final newline to terminate the file.
	exprs = append(exprs, Lit{Value: "\n"})
	return exprs, err
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

// Expr represents an expression in any language.
type Expr interface {
	fmt.Stringer
}

// Lit represents a literal string value.
type Lit struct {
	Value string
}

func (l Lit) String() string { return l.Value }
