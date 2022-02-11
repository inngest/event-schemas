package typescript

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAST(t *testing.T) {
	tests := []struct {
		Expr     []*Expr
		Expected string
	}{
		// Basic scalars
		{
			Expr: []*Expr{
				{

					Data: Scalar{
						Value: "test",
					},
				},
			},
			Expected: `"test";`,
		},
		{
			Expr: []*Expr{
				{

					Data: Scalar{
						Value: 1.1,
					},
				},
				{

					Data: Scalar{
						Value: true,
					},
				},
			},
			Expected: "1.1;\n\ntrue;",
		},
		// Local const definitions
		{
			Expr: []*Expr{
				{

					Data: Local{
						Kind:  LocalConst,
						Name:  "name",
						Value: Scalar{"Typie McTypieFace"},
					},
				},
				{

					Data: Local{
						Kind:  LocalConst,
						Name:  "isEnabled",
						Value: Scalar{true},
					},
				},
			},
			Expected: `const name = "Typie McTypieFace";

const isEnabled = true;`,
		},
		{
			Expr: []*Expr{
				{

					Data: Local{
						Kind: LocalConst,
						Name: "numbers",
						Type: strptr("number"),
						Value: Binding{
							Kind: BindingArray,
							Members: []AstKind{
								Scalar{1},
								Scalar{2},
								Scalar{3},
							},
						},
					},
				},
				{

					Data: Local{
						Kind: LocalConst,
						Name: "strings",
						Value: Binding{
							Kind: BindingArray,
							Members: []AstKind{
								Scalar{"lol"},
								Scalar{"wut"},
							},
						},
					},
				},
			},
			Expected: `const numbers: number = [1, 2, 3];

const strings = ["lol", "wut"];`,
		},
		{
			Expr: []*Expr{
				{

					Data: Local{
						Kind: LocalConst,
						Name: "obj",
						Value: Binding{
							Kind: BindingObject,
							Members: []AstKind{
								KeyValue{
									Key:   "name",
									Value: Scalar{"tester mctesty"},
								},
								KeyValue{
									Key:   "enabled",
									Value: Scalar{true},
								},
							},
						},
					},
				},
			},
			Expected: `const obj = {
  name: "tester mctesty",
  enabled: true,
};`,
		},
		// Scalar type defintions
		{
			Expr: []*Expr{
				{

					Data: Local{
						Kind:     LocalType,
						IsExport: true,
						Name:     "Name",
						Value:    Type{"string"},
					},
				},
				{

					Data: Local{
						Kind:  LocalConst,
						Name:  "myName",
						Type:  strptr("Name"),
						Value: Scalar{"coder"},
					},
				},
			},
			Expected: `export type Name = string;

const myName: Name = "coder";`,
		},
		// Interface type definitions
		{
			Expr: []*Expr{
				{

					Data: Local{
						Kind:     LocalInterface,
						IsExport: true,
						Name:     "User",
						Value: Binding{
							Kind: BindingType,
							Members: []AstKind{
								KeyValue{
									Key:   "e-mail",
									Value: Type{Value: "string"},
								},
								KeyValue{
									Key:      "name",
									Optional: true,
									Value:    Type{Value: "string"},
								},
								KeyValue{
									Key:   "loginCount",
									Value: Type{Value: "number"},
								},
								KeyValue{
									Key: "nested",
									Value: Binding{
										Kind:        BindingType,
										IndentLevel: 1,
										Members: []AstKind{
											KeyValue{
												Key:   "enabled",
												Value: Type{Value: "boolean"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Expected: `export interface User {
  "e-mail": string;
  name?: string;
  loginCount: number;
  nested: {
    enabled: boolean;
  };
};`,
		},
		// Type enum expressions
		{
			Expr: []*Expr{
				{

					Data: Enum{
						Name: "Status",
						Members: []AstKind{
							Type{"string"},
							Type{"boolean"},
						},
					},
				},
			},
			Expected: `export const Status = string | boolean;`,
		},
		// Object enums
		{
			Expr: []*Expr{
				{

					Data: Enum{
						Name: "Status",
						Members: []AstKind{
							Binding{
								Kind: BindingType,
								Members: []AstKind{
									KeyValue{
										Key:   "name",
										Value: Type{"string"},
									},
								},
							},
							Binding{
								Kind: BindingType,
								Members: []AstKind{
									KeyValue{
										Key:   "email",
										Value: Type{"string"},
									},
								},
							},
						},
					},
				},
			},
			Expected: `export const Status = {
  name: string;
} | {
  email: string;
};`,
		},
	}

	for _, test := range tests {
		actual, err := FormatAST(test.Expr...)
		require.NoError(t, err)
		require.Equal(t, test.Expected, actual)
	}
}

func strptr(s string) *string {
	return &s
}
