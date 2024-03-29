package typescript

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/inngest/event-schemas/events/marshalling"
)

const (
	indent = "  "
)

var (
	_ AstKind = Local{}
	_ AstKind = Binding{}

	constType = "const"

	alphanumeric = regexp.MustCompile(`^\w+$`)
)

func FormatAST(expr ...*Expr) (string, error) {
	str := strings.Builder{}

	for _, e := range expr {
		if _, err := str.WriteString(e.String()); err != nil {
			return "", err
		}
	}

	return str.String(), nil
}

// Expr represents a single TS expression, such as a type defintion.
//
// TODO (tonyhb): refactor.  We probably dont need Expr wrappers, and it makes
// typescript code gen function signatures a little ugly (see generateStructBinding).
// We can probably work with Local being the single top-level identifier.
type Expr struct {
	Data AstKind
}

func (e Expr) String() string {
	if _, ok := e.Data.(Lit); ok {
		return e.Data.String()
	}
	// This is code;  always add a semicolon after each expression
	str := e.Data.String()
	if strings.HasSuffix(str, ";") {
		return str
	}
	return str + ";"
}

// ASTKind is a generalization of all elements in our AST
type AstKind interface {
	isAST()
	String() string
}

func (Lit) isAST()      {}
func (Local) isAST()    {}
func (Binding) isAST()  {}
func (Enum) isAST()     {}
func (Scalar) isAST()   {}
func (Type) isAST()     {}
func (KeyValue) isAST() {}

// Lit represents literal text, such as newlines, comments, spaces, etc.
type Lit struct {
	Value string
}

func (l Lit) String() string { return l.Value }

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
	Value marshalling.Expr

	// AsType records the "as T" suffix for an identifier, eg:
	// "const Foo = 1 as int;
	AsType *string
}

func (l Local) String() string {
	var def string

	// TODO: If this is a type or an interface, capitalize the name.
	name := strings.ReplaceAll(l.Name, "#", "")

	switch l.Kind {
	case LocalConst, LocalType:
		if l.Type == nil {
			def = fmt.Sprintf("%s %s = %s", l.Kind, name, l.Value)
		} else {
			def = fmt.Sprintf("%s %s: %s = %s", l.Kind, name, *l.Type, l.Value)
		}
	case LocalInterface:
		def = fmt.Sprintf("interface %s %s", name, l.Value)
	}

	if l.AsType != nil {
		def = fmt.Sprintf("%s as %s", def, *l.AsType)
	}

	if l.IsExport {
		return fmt.Sprintf("export %s;", def)
	}

	return def

}

type BindingKind uint8

const (
	// BindingArray represents a plain ol' javascript array with contents:
	// [1, 2, 3, 4].
	BindingArray BindingKind = iota
	// BindingTypedArray represents a TS Array definition: Array<T>.  When used
	// within a Binding, any members are automatically assumed to be bound using
	// a disjunction.  However, the Members field can also contain a single
	// BindingDisjunction with many values - it does the same thing.
	//
	// Examples:
	//
	// 	Binding{
	// 		Kind: BindingTypedArray,
	// 		Members: []AstKind{
	// 			Type{"string"}, // automatically a disjunction.
	// 			Type{"number"},
	// 	     },
	// 	}
	//
	// is equivalent to:
	//
	// 	Binding{
	// 		Kind: BindingTypedArray,
	// 		Members: []AstKind{
	//			Binding{
	//				Kind: BindingDisjunction,
	//				Members: []AstKind{
	// 					Type{"string"},
	// 					Type{"number"},
	// 	     			},
	// 	     		},
	// 	     	},
	// 	}
	//
	BindingTypedArray
	// BindingObject represents a pojo.
	BindingObject
	// BindingType represents an object used for a type.  These use
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
	Members     []marshalling.Expr
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

	case BindingTypedArray:
		if len(b.Members) == 0 {
			return "Array<unknown>"
		}

		str := strings.Builder{}
		_, _ = str.WriteString("Array<")
		for n, v := range b.Members {
			_, _ = str.WriteString(v.String())
			if n < len(b.Members)-1 {
				_, _ = str.WriteString(" | ")
			}
		}
		_, _ = str.WriteString(">")
		return str.String()

	case BindingDisjunction:
		if len(b.Members) == 0 {
			return ""
		}

		str := strings.Builder{}
		for n, v := range b.Members {
			if v == nil {
				continue
			}

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
	Key      string
	Value    marshalling.Expr
	Optional bool
}

func (kv KeyValue) String() string {
	key := kv.Key
	if !alphanumeric.MatchString(kv.Key) {
		key = strconv.Quote(key)
	}

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
// In effect, this enum is a higher-kind helper for generating Typescript.  It
// creates concrete AST for enums depending on the members that it contains.
type Enum struct {
	Name    string
	Simple  bool
	Members []marshalling.Expr
}

func (e Enum) AST() ([]marshalling.Expr, error) {
	// Create a key/value AST mapping for each member of the enum.
	//
	// This allows us to dedupe values in the enum (eg. cue's int / float types
	// are just 'number';  we only want this printed once).
	kv := map[string]marshalling.Expr{}
	members := []marshalling.Expr{}

	for _, m := range e.Members {
		str := m.String()
		if _, ok := kv[str]; !ok {
			kv[str] = m
			members = append(members, m)
		}
	}

	if e.Simple {
		// This is a simple union, with no local consts.
		return []marshalling.Expr{
			Binding{
				Kind:    BindingDisjunction,
				Members: members,
			},
		}, nil
	}

	scalars := make([]marshalling.Expr, len(members))
	for n, m := range members {
		switch member := m.(type) {
		case Scalar:
			scalars[n] = KeyValue{
				Key:   strings.ToUpper(member.Unquoted()),
				Value: member,
			}
		default:
			// Immediately return a disjunction of complex types.
			return []marshalling.Expr{
				&Expr{
					Data: Local{
						Kind:     LocalType,
						Name:     e.Name,
						IsExport: true,
						Value: Binding{
							Kind: BindingDisjunction,
							// Add all members of the enum as an object.
							Members: members,
						},
					},
				},
			}, nil
		}

	}

	return []marshalling.Expr{
		&Expr{
			Data: Local{
				Kind:     LocalConst,
				Name:     e.Name,
				IsExport: true,
				AsType:   &constType,
				Value: Binding{
					Kind: BindingObject,
					// Add all members of the enum as an object.
					Members: scalars,
				},
			},
		},
		&Expr{Data: Lit{Value: "\n"}},
		&Expr{
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
	for _, item := range ast {
		str.WriteString(item.String())
	}

	if e.Simple {
		return strings.TrimSuffix(str.String(), ";")
	}
	return str.String()
}
