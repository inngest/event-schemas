package event

type Event struct {
	Name        string            `json:"name"`
	Identifiers map[string]string `json:"identifiers"`

	// Cue is the cue type definition of the event, without annotations.
	Cue string `json:"cue"`

	// Schema is the JSON schema definition of the event.
	Schema map[string]interface{} `json:"schema"`

	// Example is the canonical example event to display in the UI
	Example string `json:"example,omitempty"`
}
