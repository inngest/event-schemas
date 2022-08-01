package typescript

import (
	"context"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"github.com/inngest/event-schemas/events/marshalling"
)

var (
	ctxDepth = "indent"
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
	return marshalling.Marshal(context.Background(), v, generator{})
}

type generator struct{}

func (g generator) AST(ctx context.Context, ast []marshalling.ParsedAST) ([]marshalling.Expr, error) {
	return GenerateExprs(ctx, ast)
}

func GenerateExprs(ctx context.Context, ast []marshalling.ParsedAST) ([]marshalling.Expr, error) {
	ctx = withIncreasedDepth(ctx)
	exprs := []marshalling.Expr{}

	addExprs := func(e ...marshalling.Expr) {
		if len(exprs) > 0 && depth(ctx) == 1 {
			exprs = append(exprs, Lit{Value: "\n\n"})
		}
		exprs = append(exprs, e...)
	}

	for _, item := range ast {
		switch parsed := item.(type) {
		case *marshalling.ParsedEnum:
			next, err := generateEnum(ctx, item.(*marshalling.ParsedEnum))
			if err != nil {
				return nil, err
			}
			addExprs(next)
		case *marshalling.ParsedStruct:
			defs, err := generateStruct(ctx, item.(*marshalling.ParsedStruct))
			if err != nil {
				return nil, err
			}
			for _, def := range defs {
				addExprs(def)
			}
		case *marshalling.ParsedStructField:
			return GenerateExprs(ctx, []marshalling.ParsedAST{parsed.ParsedAST})
		case *marshalling.ParsedArray:
			defs, err := generateArray(ctx, item.(*marshalling.ParsedArray))
			if err != nil {
				return nil, err
			}
			for _, def := range defs {
				addExprs(def)
			}
		case *marshalling.ParsedIdent:
			ident := item.(*marshalling.ParsedIdent)
			addExprs(Type{Value: identToTS(ident.Ident.String())})
		case *marshalling.ParsedScalar:
			scalar := item.(*marshalling.ParsedScalar)
			addExprs(Scalar{Value: scalar.Value})
		}
	}

	return exprs, nil
}

func generateEnum(ctx context.Context, e *marshalling.ParsedEnum) (marshalling.Expr, error) {
	members, err := GenerateExprs(ctx, e.Members)
	if err != nil {
		return nil, err
	}

	var simple bool

	// If the members are basic scalars, this should be marked as 'inline' if the
	// enum is nested (eg. a struct has a field of "string" | "int", this should
	// be inline.
	if depth(ctx) > 1 {
		simple = true
		for _, m := range e.Members {
			if m.Kind() != marshalling.KindIdent {
				simple = false
			}
		}
	}

	return Enum{
		Name:    title(e.Name()),
		Simple:  simple,
		Members: members,
	}, nil
}

func generateStruct(ctx context.Context, s *marshalling.ParsedStruct) ([]marshalling.Expr, error) {
	// Structs can have embedded enums, which we pull to top-level expressions.
	binding := Binding{
		Kind:        BindingType,
		Members:     []marshalling.Expr{},
		IndentLevel: depth(ctx) - 1,
	}

	idents := []marshalling.Expr{}
	for _, member := range s.Members {
		// Generate the correct expresssions for this struct field.
		fields, err := GenerateExprs(ctx, []marshalling.ParsedAST{
			member.ParsedAST,
		})
		if err != nil {
			return nil, err
		}
		for _, field := range fields {
			switch v := field.(type) {
			case Enum:
				value := field.String()
				if !v.Simple {
					idents = append(idents, field)
					value = v.Name
				}
				// We only add this to the field if this is the
				// member name
				if v.Name != title(member.Name()) {
					continue
				}
				binding.Members = append(binding.Members, KeyValue{
					Key:      member.Name(),
					Value:    Lit{Value: value},
					Optional: member.Optional,
				})
			default:
				// This is a top-level field.
				binding.Members = append(binding.Members, KeyValue{
					Key:      member.Name(),
					Value:    field,
					Optional: member.Optional,
				})
			}
		}
	}

	exported := marshalling.Expr(binding)
	if depth(ctx) == 1 {
		// Wrap this in a definition.
		exported = Local{
			Name:     s.Name(),
			Kind:     LocalInterface,
			Value:    binding,
			IsExport: true,
		}
	}
	return append(idents, exported), nil
}

// generateArray returns an array.  This will always produce a type definition, even if all
// values in the cue list are basic literal values (eg. instead of ["1", "2"] this will generate
// Array<string>).
//
// This may return top-level expressions if the array contains a struct with enums.
func generateArray(ctx context.Context, s *marshalling.ParsedArray) ([]marshalling.Expr, error) {
	binding := Binding{
		Kind:        BindingTypedArray,
		Members:     []marshalling.Expr{},
		IndentLevel: depth(ctx) - 1,
	}

	idents := []marshalling.Expr{}

	// Arrays aren't automatically deeply nested.  Remove the depth.
	// @TODO: Remove this hack;  depth is both how deep we are within walking a nested type
	// and also how indented we are.
	level := depth(ctx) - 1
	nestedCtx := context.WithValue(ctx, ctxDepth, level)

	for _, member := range s.Members {
		// For each member, generate an expression.
		fields, err := GenerateExprs(nestedCtx, []marshalling.ParsedAST{member})
		if err != nil {
			return nil, err
		}
		for n, field := range fields {
			// We only want the last field to be added to the array;  any
			// fields prior to the last are top-level exports.
			if n < len(fields)-1 {
				idents = append(idents, field)
				continue
			}

			switch v := field.(type) {
			case Enum:
				if v.Simple {
					// A simple enum is inline.
					binding.Members = append(binding.Members, field)
					continue
				}
				// Non-simple enums are top-level.
				idents = append(idents, field)
				binding.Members = append(binding.Members, Lit{v.Name})
			case Local:
				// Local defs are always top-level
				idents = append(idents, field)
				binding.Members = append(binding.Members, Lit{v.Name})
			default:
				// This is a mmeber of the array
				binding.Members = append(binding.Members, field)
			}
		}
	}

	exported := marshalling.Expr(binding)
	if depth(ctx) == 1 {
		// Wrap this in a definition.
		exported = Local{
			Name:     s.Name(),
			Kind:     LocalType,
			Value:    binding,
			IsExport: true,
		}
	}
	return append(idents, exported), nil
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

func title(s string) string {
	s = strings.ReplaceAll(s, "#", "")
	return strings.Title(s)
}

// indentLevel returns the current indent level from the context.  This is
// a quick and dirty way of formatting nested structs.
func depth(ctx context.Context) int {
	indent, _ := ctx.Value(ctxDepth).(int)
	return indent
}

// withIncreasedIndentLevel increases the indent level in the given context,
// returning a new context with the updated indent level.
func withIncreasedDepth(ctx context.Context) context.Context {
	level := depth(ctx) + 1
	return context.WithValue(ctx, ctxDepth, level)
}
