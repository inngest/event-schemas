package fakedata

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateNumber(t *testing.T) {
	tests := []struct {
		kind  Kind
		rules []Constraint
		check func(i interface{})
	}{
		{
			kind: KindFloat,
			rules: []Constraint{
				{Rule: RuleGT, Value: 0.0},
				{Rule: RuleLTE, Value: 1.5},
			},
			check: func(i interface{}) {
				require.True(t, i.(float64) > 0)
				require.True(t, i.(float64) <= 1.5)
			},
		},
		{
			kind: KindInt,
			rules: []Constraint{
				{Rule: RuleGT, Value: 0.0},
				{Rule: RuleLT, Value: 50.0},
			},
			check: func(i interface{}) {
				require.True(t, i.(int) > 0)
				require.True(t, i.(int) < 50)
			},
		},
		{
			kind: KindInt,
			rules: []Constraint{
				{Rule: RuleGT, Value: 0.0},
				{Rule: RuleEq, Value: 29.0},
			},
			check: func(i interface{}) {
				require.True(t, i.(int) == 29)
			},
		},
		{
			kind: KindInt,
			rules: []Constraint{
				{Rule: RuleGT, Value: 0.0},
				{Rule: RuleLT, Value: 10.0},
				{Rule: RuleNEq, Value: 5.0},
				{Rule: RuleNEq, Value: 6.0},
			},
			check: func(i interface{}) {
				require.True(t, i.(int) > 0)
				require.True(t, i.(int) < 10)
				require.True(t, i.(int) != 5)
			},
		},
	}

	for _, test := range tests {
		for i := 0; i <= 100; i++ {
			rand.Seed(time.Now().UnixNano())
			val := Generate(context.Background(), test.kind, DefaultOptions, test.rules...)
			test.check(val)
		}
	}
}
