package event

type Event struct {
	Name        string                 `json:"name"`
	Identifiers map[string]string      `json:"identifiers"`
	Schema      map[string]interface{} `json:"schema"`
}
