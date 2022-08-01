package typescript

import (
	"testing"

	"github.com/inngest/event-schemas/events/marshalling"
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
				{Data: Lit{Value: "\n"}},
				{

					Data: Scalar{
						Value: true,
					},
				},
			},
			Expected: "1.1;\ntrue;",
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
				{Data: Lit{Value: "\n"}},
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
							Members: []marshalling.Expr{
								Scalar{1},
								Scalar{2},
								Scalar{3},
							},
						},
					},
				},
				{Data: Lit{Value: "\n"}},
				{

					Data: Local{
						Kind: LocalConst,
						Name: "strings",
						Value: Binding{
							Kind: BindingArray,
							Members: []marshalling.Expr{
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
		// Typed arrays
		{
			Expr: []*Expr{
				{

					Data: Local{
						Kind: LocalType,
						Name: "Contents",
						Value: Binding{
							Kind: BindingTypedArray,
							Members: []marshalling.Expr{
								Type{"string"},
								Type{"number"},
								Scalar{3.141},
							},
						},
					},
				},
			},
			Expected: `type Contents = Array<string | number | 3.141>;`,
		},
		// Objects
		{
			Expr: []*Expr{
				{

					Data: Local{
						Kind: LocalConst,
						Name: "obj",
						Value: Binding{
							Kind: BindingObject,
							Members: []marshalling.Expr{
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
				{Data: Lit{Value: "\n"}},
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
							Members: []marshalling.Expr{
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
										Members: []marshalling.Expr{
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
		// Value enums
		{
			Expr: []*Expr{
				{

					Data: Enum{
						Name: "Status",
						Members: []marshalling.Expr{
							Scalar{"push"},
							Scalar{"pull"},
						},
					},
				},
			},
			Expected: `export const Status = {
  PUSH: "push",
  PULL: "pull",
} as const;
export type Status = typeof Status[keyof typeof Status];`,
		},
		// Type enum expressions
		{
			Expr: []*Expr{
				{

					Data: Enum{
						Name: "Status",
						Members: []marshalling.Expr{
							Type{"string"},
							Type{"boolean"},
						},
					},
				},
			},
			Expected: `export type Status = string | boolean;`,
		},
		// Object enums
		{
			Expr: []*Expr{
				{

					Data: Enum{
						Name: "Status",
						Members: []marshalling.Expr{
							Binding{
								Kind: BindingType,
								Members: []marshalling.Expr{
									KeyValue{
										Key:   "name",
										Value: Type{"string"},
									},
								},
							},
							Binding{
								Kind: BindingType,
								Members: []marshalling.Expr{
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
			Expected: `export type Status = {
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
