//go:build wasm
// +build wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"cuelang.org/go/cue"
	"github.com/inngest/cuetypescript"
	"github.com/inngest/event-schemas/events/marshalling/fromjson"
	"github.com/inngest/event-schemas/events/marshalling/jsonschema"
	"github.com/inngest/event-schemas/pkg/cueutil"
	"github.com/inngest/event-schemas/pkg/merge"
)

func main() {
	// Expose the "parseCue" function
	js.Global().Set("fromJSON", js.FuncOf(FromJSON))
	js.Global().Set("toTS", js.FuncOf(ToTS))
	js.Global().Set("toJSONSchema", js.FuncOf(ToJSONSchema))
	js.Global().Set("merge", js.FuncOf(Merge))

	// To execute functions in Go you must block forever.
	<-make(chan struct{})
}

func FromJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return fmt.Sprintf("error: no JSON string provided")
	}

	input := args[0].String()
	mapped := map[string]interface{}{}
	if err := json.Unmarshal([]byte(input), &mapped); err != nil {
		return fmt.Sprintf("error parsing JSON string: %w", err)
	}

	cue, err := fromjson.FromJSON(mapped)
	if err != nil {
		return fmt.Sprintf("error generating CUE type: %w", err)
	}

	return cue
}

func Merge(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return fmt.Sprintf("error: no JSON string provided")
	}

	a := args[0].String()
	b := args[1].String()

	r := &cue.Runtime{}
	instA, err := r.Compile(".", a)
	if err != nil {
		return fmt.Sprintf("error: unable to parse a as cue: %s", err)
	}
	instB, err := r.Compile(".", b)
	if err != nil {
		return fmt.Sprintf("error: unable to parse b as cue: %s", err)
	}

	cue, err := merge.Merge(context.Background(), instA.Value(), instB.Value())
	if err != nil {
		return fmt.Sprintf("error generating CUE type: %w", err)
	}

	str, err := cueutil.ASTToSyntax(cue.Source())
	if err != nil {
		return fmt.Sprintf("error generating CUE string: %w", err)
	}

	return str
}

func ToTS(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return fmt.Sprintf("error: no cue type provided")
	}

	input := args[0].String()
	ts, err := cuetypescript.MarshalString(input)
	if err != nil {
		return fmt.Sprintf("error generating typescript: %w", err)
	}

	return ts
}

func ToJSONSchema(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return fmt.Sprintf("error: no cue type provided")
	}

	input := args[0].String()
	schemas, err := jsonschema.MarshalString(input)
	if err != nil {
		return fmt.Sprintf("error generating json schema: %w", err)
	}

	str, err := json.Marshal(schemas)
	if err != nil {
		return fmt.Sprintf("error marshalling json schema: %w", err)
	}

	return string(str)
}
