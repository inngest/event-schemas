package jsonschema

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/encoding/openapi"
)

var (
	c = &openapi.Config{
		PkgName: "",
		Version: "3.0.0",
	}
)

// MarshalString generates OpenAPI schemas given cue configuration.  Schemas are
// generated for each top-level identifier;  many schemas are generated:
//
//	#Event: {
//		name: string
//	}
//
// Cue types without identifiers will have no schemas generated.
func MarshalString(cuestr string) (Schemas, error) {
	r := &cue.Runtime{}
	inst, err := r.Compile(".", cuestr)
	if err != nil {
		return Schemas{}, fmt.Errorf("error generating inst: %w", err)
	}

	byt, err := openapi.Gen(inst, c)
	if err != nil {
		return Schemas{}, fmt.Errorf("error generating config: %w", err)
	}

	genned := &genned{}
	if err := json.Unmarshal(byt, genned); err != nil {
		return Schemas{}, fmt.Errorf("error unmarshalling genned schema: %w", err)
	}

	return Schemas{All: genned.Components.Schemas}, err
}

// MarshalCueValue generates an openAPI schema for the given cue value,
// utilizing Cue's OpenAPI integration package.  This returns a single schema
// for the given Cue value - the value must be a Cue struct containing type
// definitions.
func MarshalCueValue(v cue.Value) (map[string]interface{}, error) {
	// We need to transform the value to a *cue.Instance.
	// TODO: A bvetter way other than formatting and re-parsing to generate
	// the instance.
	val, err := formatValue(v, cue.Attributes(true))
	if err != nil {
		return nil, fmt.Errorf("error formatting instance value: %w", err)
	}

	schemas, err := MarshalString(fmt.Sprintf("#event: %s", val))
	if err != nil {
		return nil, err
	}

	return schemas.Find("event"), nil
}

// Schemas stores all schemas generated for a cue file.
type Schemas struct {
	// All stores all generated schemas, in a map.
	All map[string]map[string]interface{}
}

// Find returns a schema for the given identifier
func (s Schemas) Find(identifier string) map[string]interface{} {
	val, _ := s.All[identifier]
	return val
}

// genned represents the generated data from Cue's openapi package.  We care
// only about extracting the event schema from the generated package;  the
// rest is discarded.
type genned struct {
	Components struct {
		// Schemas lists all top-level
		Schemas map[string]map[string]interface{}
	}
}

// formatValue formats a given cue value as well-defined cue config.
func formatValue(input cue.Value, opts ...cue.Option) (string, error) {
	opts = append([]cue.Option{
		cue.Docs(true),
		cue.Optional(true),
		cue.Definitions(true),
		cue.ResolveReferences(true),
	}, opts...)

	syn := input.Syntax(opts...)
	out, err := format.Node(
		syn,
		format.TabIndent(false),
		format.UseSpaces(2),
	)
	return string(out), err
}
