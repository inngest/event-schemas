package fakedata

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"cuelang.org/go/cue"
	"github.com/stretchr/testify/require"
)

func TestFakeConstraints(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	tests := []struct {
		input       string
		constraints map[string][]Constraint
	}{
		{
			input: `{ basic: uint8 }`,
			constraints: map[string][]Constraint{
				"basic": {
					{
						Rule:  RuleGTE,
						Value: 0,
					},
					{
						Rule:  RuleLTE,
						Value: 255,
					},
				},
			},
		},
		{
			input: `{ fixed: 12 }`,
			constraints: map[string][]Constraint{
				"fixed": {
					{
						Rule:  RuleEq,
						Value: 12,
					},
				},
			},
		},
		{
			input: `{ basic: uint8 & <= 4 }`,
			constraints: map[string][]Constraint{
				"basic": {
					{
						Rule:  RuleGTE,
						Value: 0,
					},
					{
						Rule:  RuleLTE,
						Value: 255,
					},
					{
						Rule:  RuleLTE,
						Value: 4,
					},
				},
			},
		},
		{
			input: `{ basic: uint8 & (1 | 2) }`,
			constraints: map[string][]Constraint{
				"basic": {
					{
						Rule:  RuleGTE,
						Value: 0,
					},
					{
						Rule:  RuleLTE,
						Value: 255,
					},
					{
						Rule:  RuleOneOf,
						Value: []interface{}{1, 2},
					},
				},
			},
		},
		// floats
		{
			input: `{ f32: float32 }`,
			constraints: map[string][]Constraint{
				"f32": {
					{
						Rule:  RuleGTE,
						Value: math.MaxFloat32 * -1,
					},
					{
						Rule:  RuleLTE,
						Value: math.MaxFloat32,
					},
				},
			},
		},
		// nested numbers
		{
			input: `{
				data: {
					number: >= 1 & <= 10
				}
			}`,
			constraints: map[string][]Constraint{
				"data.number": {
					{
						Rule:  RuleGTE,
						Value: float64(1),
					},
					{
						Rule:  RuleLTE,
						Value: float64(10),
					},
				},
			},
		},
		{
			input: `{
				data: {
					number: int & >= 1 & <= 10
				}
			}`,
			constraints: map[string][]Constraint{
				"data.number": {
					{
						Rule:  RuleGTE,
						Value: 1,
					},
					{
						Rule:  RuleLTE,
						Value: 10,
					},
				},
			},
		},
		// strings
		{
			input: `{ str: "lol" }`,
			constraints: map[string][]Constraint{
				"str": {
					{
						Rule:  RuleEq,
						Value: "lol",
					},
				},
			},
		},
		{
			input: `{ email: string }`,
			constraints: map[string][]Constraint{
				"email": {
					{
						Rule:  RuleFormat,
						Value: FormatEmail,
					},
				},
			},
		},
	}

	for _, item := range tests {
		t.Run(item.input, func(t *testing.T) {
			result := map[string][]Constraint{}

			generatorFunc = func(ctx context.Context, kind Kind, o Options, constraints ...Constraint) interface{} {
				result[path(ctx)] = constraints
				return Generate(ctx, kind, o, constraints...)
			}

			r := &cue.Runtime{}
			inst, err := r.Compile(".", item.input)
			require.NoError(t, err)

			_, err = Fake(context.Background(), inst.Value())
			require.NoError(t, err)
			require.EqualValues(t, item.constraints, result)
		})
	}
}

func TestFakeData(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]interface{}
		seed     int64
	}{
		{
			input: `
			{
				name: "foo"
				email: string
				website_url: string
				ok: bool
				data: {
					number: >= 1 & <= 10
				}
			}
			`,
			seed: 2,
			expected: map[string]interface{}{
				"name":        "foo",
				"email":       "EJJvgRK@GlkqMrj.biz",
				"website_url": "http://www.DrHmnxY.net/fDJEeNT",
				"ok":          true,
				"data": map[string]interface{}{
					"number": 7.75,
				},
			},
		},
		{
			input: `
			{
				name: "foo"
				email: string
				website_url: string
				ok: bool
				data: {
					number: >= 1 & <= 10
				}
			}
			`,
			seed: 70211,
			expected: map[string]interface{}{
				"name":        "foo",
				"email":       "DBxgwKX@Xkjhdrh.info",
				"website_url": "https://oGlGORH.org/iyERaEO",
				"ok":          false,
				"data": map[string]interface{}{
					"number": 5.89,
				},
			},
		},
	}

	for _, test := range tests {
		r := &cue.Runtime{}
		inst, err := r.Compile(".", test.input)
		require.NoError(t, err)

		// Set the seed for deterministic testing
		DefaultOptions.Rand = rand.New(rand.NewSource(test.seed))
		output, err := Fake(context.Background(), inst.Value())

		require.NoError(t, err)
		mapped := map[string]interface{}{}
		err = output.Decode(&mapped)
		require.NoError(t, err)
		require.EqualValues(t, test.expected, mapped)
	}
}
