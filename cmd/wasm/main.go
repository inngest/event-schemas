//go:build wasm
// +build wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/inngest/cuetypescript"
	"github.com/inngest/event-schemas/events/marshalling/fromjson"
	"github.com/inngest/event-schemas/events/marshalling/jsonschema"
)

func main() {
	// Expose the "parseCue" function
	js.Global().Set("fromJSON", js.FuncOf(FromJSON))
	js.Global().Set("toTS", js.FuncOf(ToTS))
	js.Global().Set("toJSONSchema", js.FuncOf(ToJSONSchema))

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
