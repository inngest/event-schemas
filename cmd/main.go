package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/inngest/eventschema/pkg/parse"
)

func main() {
	ctx := context.Background()
	inst, err := parse.Parse(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	byt, _ := json.MarshalIndent(inst, "", "  ")
	fmt.Println(string(byt))
}
