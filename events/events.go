//go:generate go run ./internal/generate.go

package events

// Event represents a single event payload.
type Event struct {
	// Name is the unique full name of the event.
	Name string `json:"name"`

	// Version represents the version of this event.  This allows for changing
	// event schemas over time.
	Version string `json:"version"`

	// Cue is the cue type definition of the event, without annotations.
	Cue string `json:"cue"`

	// Schema is the JSON schema definition of the event.
	Schema map[string]interface{} `json:"schema"`

	// Example is the canonical example event to display in the UI
	Example string `json:"example,omitempty"`
}
