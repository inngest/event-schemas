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
					name:  "#Def",
					Value: "scalar",
				},
			},
		},
		{
			name:  "string ident",
			input: `#MyVar: string`,
			expected: []ParsedAST{
				&ParsedIdent{
					name:  "#MyVar",
					Ident: ast.NewIdent("string"),
				},
			},
		},
		{
			name:  "int ident",
			input: `#Def: int`,
			expected: []ParsedAST{
				&ParsedIdent{
					name:  "#Def",
					Ident: ast.NewIdent("int"),
				},
			},
		},
		{
			name:  "int ident with constraints",
			input: `#Def: int & >= 5 & <= 10`,
			expected: []ParsedAST{
				&ParsedIdent{
					name:  "#Def",
					Ident: ast.NewIdent("int"),
				},
			},
		},
		{
			name:  "int ident with constraints and default",
			input: `#Def: int & >= 5 & <= 10 | *8`,
			expected: []ParsedAST{
				&ParsedIdent{
					name:  "#Def",
					Ident: ast.NewIdent("int"),
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
					name: "#Person",
					Members: []*ParsedStructField{
						{
							ParsedAST: &ParsedIdent{
								name:  "name",
								Ident: ast.NewIdent("string"),
							},
						},
						{
							ParsedAST: &ParsedIdent{
								name:  "age",
								Ident: ast.NewIdent("int"),
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
					name: "#Nested",
					Members: []*ParsedStructField{
						{
							ParsedAST: &ParsedStruct{
								name: "nested",
								Members: []*ParsedStructField{
									{
										ParsedAST: &ParsedEnum{
											name: "enum",
											Members: []ParsedAST{
												&ParsedScalar{Value: "test"},
												&ParsedScalar{Value: "another"},
											},
										},
									},
									{
										ParsedAST: &ParsedEnum{
											name: "types",
											Members: []ParsedAST{
												&ParsedIdent{Ident: ast.NewIdent("string")},
												&ParsedIdent{Ident: ast.NewIdent("int")},
											},
										},
									},
									{
										ParsedAST: &ParsedIdent{
											name:  "opt",
											Ident: ast.NewIdent("string"),
										},
										Optional: true,
									},
									{
										ParsedAST: &ParsedStruct{
											name: "some",
											Members: []*ParsedStructField{
												{
													ParsedAST: &ParsedIdent{
														name:  "item",
														Ident: ast.NewIdent("bool"),
													},
												},
											},
										},
									},
								},
							},
						},
						{
							ParsedAST: &ParsedIdent{
								name:  "title",
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
					name: "#Types",
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
								Ident: ast.NewIdent("string"),
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
					name: "#Idents",
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
