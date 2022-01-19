package parse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/openapi"
	"github.com/inngest/eventschema/defs"
	"github.com/inngest/eventschema/events"
)

var (
	c = &openapi.Config{
		PkgName: "",
		Version: "3.0.0",
	}
)

// Parse evaluates all embeded cue files within defs/cue.mod, returning parsed event
// information from the cue types.
func Parse(ctx context.Context) ([]events.Event, error) {
	insts, err := instances(ctx)
	if err != nil {
		return nil, err
	}

	events := []events.Event{}

	for _, i := range insts {
		// Iterate through each value within the instance (file) and parse
		// the event.
		e, err := walkDefinitions(i.Value())
		if err != nil {
			return nil, err
		}
		events = append(events, e...)
	}

	return events, nil
}

// instances parses all embeded cue files, returning cue Instances representing each
// file.
func instances(ctx context.Context) ([]*cue.Instance, error) {
	instances := []*cue.Instance{}

	r := &cue.Runtime{}
	cfg := load.Config{
		Overlay:    map[string]load.Source{},
		Dir:        "/cue.mod/",
		ModuleRoot: "/",
		Package:    "*",
		Stdin:      bytes.NewBuffer(nil),
	}

	err := fs.WalkDir(defs.FS, ".", func(p string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		contents, err := defs.FS.ReadFile(p)
		if err != nil {
			return err
		}

		cfg.Overlay[path.Join("/", p)] = load.FromBytes(contents)
		return nil
	})

	if err != nil {
		return nil, err
	}

	bis := load.Instances([]string{""}, &cfg)
	for _, i := range bis {
		if i.Err != nil {
			return nil, fmt.Errorf("error loading instance: %w", i.Err)
		}

		inst, err := r.Build(i)
		if err != nil {
			return nil, fmt.Errorf("error building instance: %w", err)
		}

		instances = append(instances, inst)
	}

	return instances, nil
}

// walkDefinitions walks through each definition within a Cue instance, finds
// every definition that contains an event schema, then parses the event schema
// from the Cue type definition.
func walkDefinitions(v cue.Value) ([]events.Event, error) {
	events := []events.Event{}

	it, err := v.Fields()
	if err != nil {
		return nil, err
	}

	for it.Next() {
		if it.IsDefinition() {
			continue
		}

		val := it.Value()
		// If the value has a field "schema", it's part of our definition.
		if _, err := val.LookupField("schema"); err != nil {
			continue
		}

		evt, err := gen(val)
		if err != nil {
			return nil, err
		}

		events = append(events, *evt)

	}

	return events, nil
}

// gen generates a new event given the event definition as a cue.Value.
func gen(v cue.Value) (*events.Event, error) {
	// If the value has a field "schema", it's part of our definition.
	sf, err := v.LookupField("schema")
	if err != nil {
		return nil, fmt.Errorf("unable to find schema field")
	}
	schema, err := schema(sf.Value)
	if err != nil {
		return nil, err
	}

	cuedef, _ := formatValue(sf.Value, cue.Attributes(false))

	name := cueString(sf.Value, "name")

	evt := &events.Event{
		Name:    name,
		Schema:  schema,
		Cue:     cuedef,
		Example: cueString(v, "example"),
		Version: cueString(sf.Value, "v"),
	}

	return evt, nil
}

// schema generates an openAPI schema for the given schema field of an event,
// utilizing Cue's OpenAPI integration package.
//
// This should be called for a single event instance.
func schema(v cue.Value) (map[string]interface{}, error) {
	val, err := formatValue(v, cue.Attributes(true))
	if err != nil {
		return nil, fmt.Errorf("error formatting instance value: %w", err)
	}

	r := &cue.Runtime{}
	inst, err := r.Compile(".", fmt.Sprintf("#event: %s", val))
	if err != nil {
		return nil, fmt.Errorf("error generating inst: %w", err)
	}

	byt, err := openapi.Gen(inst, c)
	if err != nil {
		return nil, fmt.Errorf("error generating config: %w", err)
	}

	genned := &genned{}
	if err := json.Unmarshal(byt, genned); err != nil {
		return nil, fmt.Errorf("error unmarshalling genned schema: %w", err)
	}

	return genned.Components.Schemas.Event, err
}

// genned represents the generated data from Cue's openapi package.  We care
// only about extracting the event schema from the generated package;  the
// rest is discarded.
type genned struct {
	Components struct {
		Schemas struct {
			Event map[string]interface{}
		}
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
