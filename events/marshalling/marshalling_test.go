package marshalling

import (
	"context"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ParsedAST
		err      error
	}{
		// scalars
		{
			name:  "basic",
			input: `#Def: "scalar"`,
			expected: []ParsedAST{
				&ParsedScalar{
					Name:  "#Def",
					Value: "scalar",
				},
			},
		},
		{
			name:  "string ident",
			input: `#MyVar: string`,
			expected: []ParsedAST{
				&ParsedIdent{
					Name: "#MyVar",
					Ident: &ast.Ident{
						Name: "string",
					},
				},
			},
		},
		{
			name:  "int ident",
			input: `#Def: int`,
			expected: []ParsedAST{
				&ParsedIdent{
					Name: "#Def",
					Ident: &ast.Ident{
						Name: "int",
					},
				},
			},
		},
		{
			name:  "int ident with constraints",
			input: `#Def: int & >= 5 & <= 10`,
			expected: []ParsedAST{
				&ParsedIdent{
					Name: "#Def",
					Ident: &ast.Ident{
						Name: "int",
					},
				},
			},
		},
		{
			name:  "int ident with constraints and default",
			input: `#Def: int & >= 5 & <= 10 | *8`,
			expected: []ParsedAST{
				&ParsedIdent{
					Name: "#Def",
					Ident: &ast.Ident{
						Name: "int",
					},
					Default: &ParsedScalar{
						Value: 8,
					},
				},
			},
		},
		// structs
		{
			name: "basic struct",
			input: `#Person: {
			name: string,
			age: uint & >= 0 | *21,
			}`,
			expected: []ParsedAST{
				&ParsedStruct{
					Name: "#Person",
					Members: map[string]ParsedAST{
						"name": &ParsedIdent{
							Name: "name",
							Ident: &ast.Ident{
								Name: "string",
							},
						},
						"age": &ParsedIdent{
							Name: "age",
							Ident: &ast.Ident{
								Name: "int",
							},
							Default: &ParsedScalar{
								Value: 21,
							},
						},
					},
				},
			},
		},
		// arrays
		{
			name:  "basic array",
			input: `#Types: [...string | int | float64] | *[string]`,
			expected: []ParsedAST{
				&ParsedArray{
					Name: "#Types",
					Members: []ParsedAST{
						&ParsedEnum{
							Members: []ParsedAST{
								&ParsedIdent{
									Ident: ast.NewIdent("string"),
								},
								&ParsedIdent{
									Ident: ast.NewIdent("int"),
								},
								&ParsedIdent{
									Ident: ast.NewIdent("float64"),
								},
							},
						},
					},
					Default: &ParsedArray{
						Members: []ParsedAST{
							&ParsedIdent{
								Ident: &ast.Ident{
									Name: "string",
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "lit array",
			input: `#Idents: ["person", "dog", "cat"]`,
			expected: []ParsedAST{
				&ParsedArray{
					Name: "#Idents",
					Members: []ParsedAST{
						&ParsedScalar{
							Value: "person",
						},
						&ParsedScalar{
							Value: "dog",
						},
						&ParsedScalar{
							Value: "cat",
						},
					},
				},
			},
		},
	}

	for _, n := range tests {
		test := n
		t.Run(test.name, func(t *testing.T) {
			r := &cue.Runtime{}
			inst, err := r.Compile(".", test.input)
			require.NoError(t, err)
			actual, err := Parse(context.Background(), inst.Value())
			require.NoError(t, err)
			require.EqualValues(t, test.expected, actual)
		})
	}
}
