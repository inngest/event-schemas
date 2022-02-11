package typescript

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	indent = "  "
)

var (
	_ AstKind = Local{}
	_ AstKind = Binding{}

	alphanumeric = regexp.MustCompile(`^\w+$`)
)

func FormatAST(expr ...*Expr) (string, error) {
	str := strings.Builder{}

	for n, e := range expr {
		if _, err := str.WriteString(e.String()); err != nil {
			return "", err
		}
		if n < len(expr)-1 {
			if _, err := str.WriteString("\n\n"); err != nil {
				return "", err
			}
		}
	}

	return str.String(), nil
}

// Expr represents a single TS expression, such as a type defintion.
type Expr struct {
	Data AstKind
}

func (e Expr) String() string {
	return e.Data.String() + ";"
}

// ASTKind is a generalization of all elements in our AST
type AstKind interface {
	isAST()
	String() string
}

func (Local) isAST()    {}
func (Binding) isAST()  {}
func (Enum) isAST()     {}
func (Scalar) isAST()   {}
func (Type) isAST()     {}
func (KeyValue) isAST() {}

// Scalar represents a scalar value such as a string, number, boolean,
// etc.
type Scalar struct {
	Value interface{}
}

func (s Scalar) String() string {
	switch t := s.Value.(type) {
	case string:
		return strconv.Quote(t)
	default:
		return fmt.Sprintf("%v", s.Value)
	}
}

func (s Scalar) Unquoted() string {
	return fmt.Sprintf("%v", s.Value)
}

type Type struct {
	Value string
}

func (t Type) String() string {
	return t.Value
}

// LocalKind represents a local variable or type definition
type LocalKind uint8

const (
	LocalConst LocalKind = iota
	LocalType
	LocalInterface
)

func (l LocalKind) String() string {
	switch l {
	case LocalConst:
		return "const"
	case LocalType:
		return "type"
	case LocalInterface:
		return "interface"
	}
	return ""
}

// Local is a concrete AST expr which represents a definition
// for a variable or type.
type Local struct {
	// Kind is the type of definition for the variable.
	Kind LocalKind
	// Name represents the name for the identifier
	Name string
	// Type represents the optional type definition for the identifier.
	// This is valid for const definitions only.
	Type *string
	// IsExport defines whether this identifier should be exported.
	IsExport bool
	// Value is the value that this identifier refers to.  This could be
	// a scalar, a type, a binding, etc.
	Value AstKind
}

func (l Local) String() string {
	var def string

	// TODO: If this is a type or an interface, capitalize the name.
	switch l.Kind {
	case LocalConst, LocalType:
		if l.Type == nil {
			def = fmt.Sprintf("%s %s = %s", l.Kind, l.Name, l.Value)
		} else {
			def = fmt.Sprintf("%s %s: %s = %s", l.Kind, l.Name, *l.Type, l.Value)
		}
	case LocalInterface:
		// TODO: Create interface
		def = fmt.Sprintf("interface %s %s", l.Name, l.Value.String())
	}

	if l.IsExport {
		return fmt.Sprintf("export %s", def)
	}

	return def

}

type BindingKind uint8

const (
	BindingArray BindingKind = iota
	BindingObject
	// Type represents an object used for a type.  These use
	// semi-colons as their field terminators.
	BindingType
	// BindingDisjunction represents an ADT enum - values combined
	// with " | "
	BindingDisjunction
)

// Binding represents a complex type: an array, enum, object, etc.
type Binding struct {
	IndentLevel int
	Kind        BindingKind
	Members     []AstKind
}

func (b Binding) String() string {

	switch b.Kind {
	case BindingArray:
		if len(b.Members) == 0 {
			return "[]"
		}

		str := strings.Builder{}
		_, _ = str.WriteString("[")
		for n, v := range b.Members {
			_, _ = str.WriteString(v.String())
			if n < len(b.Members)-1 {
				_, _ = str.WriteString(", ")
			}
		}
		_, _ = str.WriteString("]")
		return str.String()

	case BindingDisjunction:
		if len(b.Members) == 0 {
			return ""
		}

		str := strings.Builder{}
		for n, v := range b.Members {
			_, _ = str.WriteString(v.String())
			if n < len(b.Members)-1 {
				_, _ = str.WriteString(" | ")
			}
		}
		return str.String()

	case BindingObject, BindingType:
		if len(b.Members) == 0 {
			return "{}"
		}

		term := ","
		if b.Kind == BindingType {
			term = ";"
		}

		str := strings.Builder{}

		_, _ = str.WriteString("{\n")

		for _, v := range b.Members {
			for i := 0; i <= b.IndentLevel; i++ {
				_, _ = str.WriteString(indent)
			}
			_, _ = str.WriteString(fmt.Sprintf("%s%s\n", v.String(), term))
		}

		// Add indents to the terminator, eg.
		// {
		//    foo: {
		//      bar: true,
		//    } // this indent
		// }
		for i := 0; i < b.IndentLevel; i++ {
			_, _ = str.WriteString(indent)
		}
		_, _ = str.WriteString("}")

		return str.String()
	}

	return ""
}

// KeyValue represents a key and value within a BindingObject or Enum
type KeyValue struct {
	Key   string
	Value AstKind

	Optional bool
}

func (kv KeyValue) String() string {
	key := kv.Key
	if !alphanumeric.MatchString(kv.Key) {
		key = strconv.Quote(key)
	}

	// TODO: Does this key have non-alpha characters?
	if kv.Optional {
		return fmt.Sprintf("%s?: %s", key, kv.Value.String())
	}
	return fmt.Sprintf("%s: %s", key, kv.Value.String())
}

// An Enum is an ADT - a union type within Cue.  We special-case enums because
// of typescript limitations. A pure `enum Foo {...}` in typescript isn't that
// fun to use;  it's recommended by many to create an Object containing the enum
// values, then use `typeof EnumName[keyof typeof EnumName]` to define the enum.
//
// We have multiple different enums available.  A cue enum could be `1 | 2 | 3`.
type Enum struct {
	Name    string
	Members []AstKind
}

func (e Enum) AST() ([]*Expr, error) {
	// Create a key/value AST mapping for each member of the enum.
	kv := make([]AstKind, len(e.Members))

	for n, m := range e.Members {
		switch member := m.(type) {
		case Scalar:
			kv[n] = KeyValue{
				Key:   strings.ToUpper(member.Unquoted()),
				Value: member,
			}
		default:
			// Immediately return a disjunction of complex types.
			return []*Expr{
				{
					Data: Local{
						Kind:     LocalConst,
						Name:     e.Name,
						IsExport: true,
						Value: Binding{
							Kind: BindingDisjunction,
							// Add all members of the enum as an object.
							Members: e.Members,
						},
					},
				},
			}, nil
		}

	}

	return []*Expr{
		{
			Data: Local{
				Kind:     LocalConst,
				Name:     e.Name,
				IsExport: true,
				Value: Binding{
					Kind: BindingObject,
					// Add all members of the enum as an object.
					Members: kv,
				},
			},
		},
		{
			Data: Local{
				Kind:     LocalType,
				Name:     e.Name,
				IsExport: true,
				Value: Type{
					Value: fmt.Sprintf("typeof %s[keyof typeof %s]", e.Name, e.Name),
				},
			},
		},
	}, nil
}

func (e Enum) String() string {
	ast, err := e.AST()
	if err != nil {
		return err.Error()
	}
	str := strings.Builder{}
	for n, item := range ast {
		str.WriteString(item.String())

		if n < len(ast)-1 {
			str.WriteString("\n")
		}
	}
	return strings.TrimSuffix(str.String(), ";")
}
