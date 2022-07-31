package marshalling

import "cuelang.org/go/cue/ast"

// ParsedAST is an interface which each parsed AST fulfills.  This lets
// us specify an interface for eg. struct or enum member values.
type ParsedAST interface {
	// Kind represents the AST kind that the parsed AST represents.
	Kind() string
	// SetDefault sets the default value, if found.
	SetDefault(to interface{})
}

type ParsedEnum struct {
	Name    string
	Members []ParsedAST
	Default interface{}
}

func (ParsedEnum) Kind() string { return "enum" }

func (p *ParsedEnum) SetDefault(to interface{}) {
	p.Default = to
}

type ParsedStruct struct {
	Name    string
	Members map[string]ParsedAST
	Default interface{}
}

func (ParsedStruct) Kind() string { return "struct" }

func (p *ParsedStruct) SetDefault(to interface{}) {
	p.Default = to
}

// ParsedArray represents an array type sepcified within Cue.
//
// A Cue array can contain many different types;  it is not constrained
// to storing a single type.  Due to this, we store all available parsed
// types within the Types field.
type ParsedArray struct {
	Name    string
	Members []ParsedAST
	Default interface{}
}

func (ParsedArray) Kind() string { return "array" }

func (p *ParsedArray) SetDefault(to interface{}) {
	p.Default = to
}

// ParsedIdent represents a single scalar type, eg. a string or a number
type ParsedIdent struct {
	Name string
	// Type represnets the ast Ident that was parsed.
	Ident *ast.Ident

	// Default represents the default value, if any.
	Default interface{}

	// TODO: Expose constraints.
	constraints []interface{}
}

func (ParsedIdent) Kind() string { return "ident" }

func (p *ParsedIdent) SetDefault(to interface{}) {
	p.Default = to
}

// ParsedScalar represents a single concrete scalar value, eg. a string instance
// "foo" or a number instance 42.
type ParsedScalar struct {
	Name    string
	Value   interface{}
	Default interface{}
}

func (ParsedScalar) Kind() string { return "scalar" }

// SetDefault is a no-op with scalars, as they're concrete.
func (*ParsedScalar) SetDefault(to interface{}) {
	panic("impossible on scalars")
}
