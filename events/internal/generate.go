//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/inngest/event-schemas/events/internal/parse"
)

func main() {
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

	// Write the JSON file to events/generated.json
	if err := os.WriteFile("generated.json", byt, 0600); err != nil {
		return err
	}

	return nil
}
