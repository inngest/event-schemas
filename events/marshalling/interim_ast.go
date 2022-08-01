package marshalling

import "cuelang.org/go/cue/ast"

type ParsedKind int

const (
	KindNone ParsedKind = iota
	KindEnum
	KindStruct
	KindArray
	KindIdent
	KindScalar
)

// ParsedAST is an interface which each parsed AST fulfills.  This lets
// us specify an interface for eg. struct or enum member values.
type ParsedAST interface {
	// Kind represents the AST kind that the parsed AST represents.
	Kind() ParsedKind
	// name returns the name for the field, if specified.
	Name() string
	// SetDefault sets the default value, if found.
	SetDefault(to interface{})
}

type ParsedEnum struct {
	name    string
	Members []ParsedAST
	Default interface{}
}

func (ParsedEnum) Kind() ParsedKind { return KindEnum }

func (p ParsedEnum) Name() string { return p.name }

func (p *ParsedEnum) SetDefault(to interface{}) {
	p.Default = to
}

type ParsedStruct struct {
	Members []*ParsedStructField
	Default interface{}

	name string
}

func (ParsedStruct) Kind() ParsedKind { return KindStruct }

func (p ParsedStruct) Name() string { return p.name }

func (p *ParsedStruct) SetDefault(to interface{}) {
	p.Default = to
}

type ParsedStructField struct {
	ParsedAST
	Optional bool
}

// ParsedArray represents an array type sepcified within Cue.
//
// A Cue array can contain many different types;  it is not constrained
// to storing a single type.  Due to this, we store all available parsed
// types within the Types field.
type ParsedArray struct {
	name     string
	Members  []ParsedAST
	Default  interface{}
	Optional bool
}

func (ParsedArray) Kind() ParsedKind { return KindArray }

func (p ParsedArray) Name() string { return p.name }

func (p *ParsedArray) SetDefault(to interface{}) {
	p.Default = to
}

// ParsedIdent represents a single scalar type, eg. a string or a number
type ParsedIdent struct {
	name string

	// Type represnets the ast Ident that was parsed.
	Ident *ast.Ident

	// Default represents the default value, if any.
	Default interface{}

	// TODO: Expose constraints.
	constraints []interface{}
}

func (ParsedIdent) Kind() ParsedKind { return KindIdent }

func (p ParsedIdent) Name() string { return p.name }

func (p *ParsedIdent) SetDefault(to interface{}) {
	p.Default = to
}

// ParsedScalar represents a single concrete scalar value, eg. a string instance
// "foo" or a number instance 42.
type ParsedScalar struct {
	name     string
	Value    interface{}
	Default  interface{}
	Optional bool
}

func (ParsedScalar) Kind() ParsedKind { return KindScalar }

func (p ParsedScalar) Name() string { return p.name }

// SetDefault is a no-op with scalars, as they're concrete.
func (*ParsedScalar) SetDefault(to interface{}) {
	panic("impossible on scalars")
}
