package parse

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/openapi"
	"github.com/inngest/cuetypescript"
	"github.com/inngest/event-schemas/defs"
	"github.com/inngest/event-schemas/events"
	"github.com/inngest/event-schemas/events/marshalling/jsonschema"
)

const (
	serviceSeparator = "/"

	typescriptEventName = "InngestEvent"
)

var (
	c = &openapi.Config{
		PkgName: "",
		Version: "3.0.0",
	}
	nonAlphaRegexp = regexp.MustCompile("[^\\w]|_")
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
		e, err := walkDefinitions(i.Value(), i)
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
func walkDefinitions(v cue.Value, i *cue.Instance) ([]events.Event, error) {
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

		evt, err := gen(val, i)
		if err != nil {
			return nil, err
		}

		events = append(events, *evt)

	}

	return events, nil
}

// gen generates a new event given the event definition as a cue.Value.
func gen(v cue.Value, i *cue.Instance) (*events.Event, error) {
	// If the value has a field "schema", it's part of our definition.
	sf, err := v.LookupField("schema")
	if err != nil {
		return nil, fmt.Errorf("unable to find schema field")
	}

	// TODO: Recursively walk the definition and see if there are any references
	// to other cue definitions.  If so, we need to create a new instance which
	// contains every definition to be marshalled, else the schema field will have
	// references to undefined definitions.
	//
	// It would be good to be able to pass in the entire *cue.Instance of our
	// objects, but that's not possible as we're attempting to create a JSON schema
	// for a field in a struct _and_ JSON-schema requires a strict format to parse
	// an instance.

	schema, err := jsonschema.MarshalCueValue(sf.Value)
	if err != nil {
		return nil, err
	}

	cuedef, _ := formatValue(sf.Value, cue.Attributes(false))
	name := cueString(sf.Value, "name")

	// Marshal the typescript in an embedded event.
	ts, err := genTypescript(name, sf.Value)
	if err != nil {
		return nil, err
	}

	var service string
	parts := strings.SplitN(name, serviceSeparator, 2)
	if len(parts) == 2 {
		service = parts[0]
	}

	examples := []map[string]interface{}{}
	err = cueField(v, "examples").Value.Decode(&examples)

	evt := &events.Event{
		Name:        name,
		Service:     service,
		Description: cueString(v, "description"),
		Schema:      schema,
		TypeScript:  ts,
		Cue:         cuedef,
		Examples:    examples,
		Version:     cueString(sf.Value, "v"),
	}

	return evt, nil
}

func genTypescript(name string, v cue.Value) (string, error) {
	// Create a new struct which wraps the cue schema.
	val, err := formatValue(v, cue.Attributes(true))
	if err != nil {
		return "", fmt.Errorf("error formatting schema value: %w", err)
	}
	r := &cue.Runtime{}
	inst, err := r.Compile(".", fmt.Sprintf("%s: %s", typescriptEventName, val))
	if err != nil {
		return "", fmt.Errorf("error wrapping schema with event name: %w", err)
	}

	return cuetypescript.MarshalCueValue(inst.Value())
}

func titleCaseName(name string) string {
	name = nonAlphaRegexp.ReplaceAllString(name, " ")
	name = strings.Title(name)
	return strings.ReplaceAll(name, " ", "")
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
