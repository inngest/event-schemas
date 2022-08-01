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
					Members: map[string]*ParsedStructField{
						"name": &ParsedStructField{
							ParsedAST: &ParsedIdent{
								Name: "name",
								Ident: &ast.Ident{
									Name: "string",
								},
							},
						},
						"age": &ParsedStructField{
							ParsedAST: &ParsedIdent{
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
		},
		{
			name: "basic struct",
			input: `#Nested: {
			nested: {
				enum: "test" | "another"
				types: string | int
				opt?: string
				some: {
					item: bool
				}
			}
			title: string,
			}`,
			expected: []ParsedAST{
				&ParsedStruct{
					Name: "#Nested",
					Members: map[string]*ParsedStructField{
						"nested": &ParsedStructField{
							ParsedAST: &ParsedStruct{
								Name: "nested",
								Members: map[string]*ParsedStructField{
									"enum": &ParsedStructField{
										ParsedAST: &ParsedEnum{
											Name: "enum",
											Members: []ParsedAST{
												&ParsedScalar{Value: "test"},
												&ParsedScalar{Value: "another"},
											},
										},
									},
									"types": &ParsedStructField{
										ParsedAST: &ParsedEnum{
											Name: "types",
											Members: []ParsedAST{
												&ParsedIdent{Ident: ast.NewIdent("string")},
												&ParsedIdent{Ident: ast.NewIdent("int")},
											},
										},
									},
									"opt": &ParsedStructField{
										ParsedAST: &ParsedIdent{
											Name:  "opt",
											Ident: ast.NewIdent("string"),
										},
										Optional: true,
									},
									"some": &ParsedStructField{
										ParsedAST: &ParsedStruct{
											Name: "some",
											Members: map[string]*ParsedStructField{
												"item": &ParsedStructField{
													ParsedAST: &ParsedIdent{
														Name:  "item",
														Ident: ast.NewIdent("bool"),
													},
												},
											},
										},
									},
								},
							},
						},
						"title": &ParsedStructField{
							ParsedAST: &ParsedIdent{
								Name:  "title",
								Ident: ast.NewIdent("string"),
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
