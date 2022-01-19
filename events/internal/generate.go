//go:build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/inngest/eventschema/events"
	"github.com/inngest/eventschema/events/internal/parse"
)

func main() {
	fmt.Printf("%#v\n", events.Events)

	if err := generateJSON(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func generateJSON() error {
	ctx := context.Background()
	events, err := parse.Parse(ctx)
	if err != nil {
		return err
	}

	// XXX: We can use a fast marshaller here, such as fastjson, as
	// we know the event shape already.
	byt, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}

	t, err := template.New("go").Parse(tpl)
	if err != nil {
		return err
	}

	var code bytes.Buffer
	err = t.Execute(&code, map[string]interface{}{
		"Encoded": strconv.Quote(string(byt)),
	})
	if err != nil {
		return err
	}

	return os.WriteFile("generated.go", code.Bytes(), 0600)
}

const tpl = `// Code generated by generate.go; DO NOT EDIT.

package events

import (
	"encoding/json"
)

// Events list a subset of events ingested via integrations into Inngest,
// with cue and JSON schema fields documenting their format.
var Events []Event

func init() {
	_ = json.Unmarshal([]byte(encoded), &Events)
}

// encoded stores the JSON encoded event struct with pre-parsed cue and JSON
// schema definitions.
const encoded = {{ .Encoded }}

`
