package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalString(t *testing.T) {
	input := `{
      "properties": {
        "data": {
          "description": "The event payload, containing all event data",
          "properties": {
            "context": {
              "properties": {
                "ip": {
                  "type": "string"
                },
                "ua": {
                  "type": "string"
                }
              },
              "required": [
                "ua",
                "ip"
              ],
              "type": "object",
              "additionalProperties": false
            },
            "created_at": {
              "type": "string"
            },
            "email": {
              "type": "string"
            }
          },
          "required": [
            "email"
          ],
          "type": "object",
          "additionalProperties": false
        },
        "name": {
          "description": "The unique name of the event",
          "enum": [
            "test.event"
          ],
          "type": "string"
        },
        "ts": {
          "description": "The epoch of the event, in milliseconds",
          "type": "number"
        },
        "user": {
          "description": "User information for the author of the event",
          "properties": {
            "email": {
              "type": "string"
            },
            "external_id": {
              "type": "string"
            }
          },
          "required": [
            "external_id",
            "email"
          ],
          "type": "object",
          "additionalProperties": false
        },
        "v": {
          "description": "An optional event version",
          "enum": [
            "1"
          ],
          "type": "string"
        }
      },
      "required": [
        "name",
        "data",
        "user",
        "v"
      ],
      "type": "object",
      "additionalProperties": false
    }`

	expected := `{
  // The event payload, containing all event data
  data: {
    context?: {
      ip: string
      ua: string
    }
    created_at?: string
    email:       string
  }

  // The unique name of the event
  name: "test.event"

  // The epoch of the event, in milliseconds
  ts?: number

  // User information for the author of the event
  user: {
    email:       string
    external_id: string
  }

  // An optional event version
  v: "1"
}`

	// Assert that unmarshalling a JSON schema object is valid.
	actual, err := UnmarshalString(input)
	require.NoError(t, err)
	require.Equal(t, expected, actual, "Received:\n%s\n", actual)

}
