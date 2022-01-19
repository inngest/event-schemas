// Code generated by generate.go; DO NOT EDIT.

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
const encoded = "[\n  {\n    \"name\": \"github/pull_request\",\n    \"integration\": \"github\",\n    \"description\": \"Created when pull requests are created or modified\",\n    \"version\": \"\",\n    \"cue\": \"{\\n  // The unique name of the event\\n  name: \\\"github/pull_request\\\"\\n  // The event payload, containing all event data\\n  data: {\\n    action: \\\"opened\\\" | \\\"closed\\\" | \\\"merged\\\" | \\\"review_requested\\\" | \\\"synchronize\\\" | \\\"edited\\\"\\n\\n    // The pull request number.  Also contained within pull_request\\n    number: \\u003e=1\\n    pull_request: {\\n      // The pull request number.  Also specified within  top-level data\\n      number: \\u003e=1\\n\\n      // The pull request title\\n      title: string\\n\\n      // The pull request description\\n      body: string\\n\\n      // The number of individual commits wanting to be merged\\n      commits: \\u003e=1\\n\\n      // The number of changed files\\n      changed_files: \\u003e=1\\n\\n      // Whether the pull request is a draft\\n      draft: bool\\n    }\\n\\n    // The commit hash of the tip of the PR before changes\\n    before?: string\\n    // The commit hash of the tip of the PR after changes\\n    after?: string\\n  }\\n  // User information for the author of the event\\n\\n  // There is no user information available within this event.\\n  user: {}\\n\\n  // An optional event version\\n  v?: string\\n}\",\n    \"schema\": {\n      \"properties\": {\n        \"data\": {\n          \"description\": \"The event payload, containing all event data\",\n          \"properties\": {\n            \"action\": {\n              \"enum\": [\n                \"opened\",\n                \"closed\",\n                \"merged\",\n                \"review_requested\",\n                \"synchronize\",\n                \"edited\"\n              ],\n              \"type\": \"string\"\n            },\n            \"after\": {\n              \"description\": \"The commit hash of the tip of the PR after changes\",\n              \"type\": \"string\"\n            },\n            \"before\": {\n              \"description\": \"The commit hash of the tip of the PR before changes\",\n              \"type\": \"string\"\n            },\n            \"number\": {\n              \"description\": \"The pull request number.  Also contained within pull_request\",\n              \"minimum\": 1,\n              \"type\": \"number\"\n            },\n            \"pull_request\": {\n              \"properties\": {\n                \"body\": {\n                  \"description\": \"The pull request description\",\n                  \"type\": \"string\"\n                },\n                \"changed_files\": {\n                  \"description\": \"The number of changed files\",\n                  \"minimum\": 1,\n                  \"type\": \"number\"\n                },\n                \"commits\": {\n                  \"description\": \"The number of individual commits wanting to be merged\",\n                  \"minimum\": 1,\n                  \"type\": \"number\"\n                },\n                \"draft\": {\n                  \"description\": \"Whether the pull request is a draft\",\n                  \"type\": \"boolean\"\n                },\n                \"number\": {\n                  \"description\": \"The pull request number.  Also specified within  top-level data\",\n                  \"minimum\": 1,\n                  \"type\": \"number\"\n                },\n                \"title\": {\n                  \"description\": \"The pull request title\",\n                  \"type\": \"string\"\n                }\n              },\n              \"required\": [\n                \"number\",\n                \"title\",\n                \"body\",\n                \"commits\",\n                \"changed_files\",\n                \"draft\"\n              ],\n              \"type\": \"object\"\n            }\n          },\n          \"required\": [\n            \"action\",\n            \"number\",\n            \"pull_request\"\n          ],\n          \"type\": \"object\"\n        },\n        \"name\": {\n          \"description\": \"The unique name of the event\",\n          \"enum\": [\n            \"github/pull_request\"\n          ],\n          \"type\": \"string\"\n        },\n        \"user\": {\n          \"description\": \"There is no user information available within this event.\",\n          \"type\": \"object\"\n        },\n        \"v\": {\n          \"description\": \"An optional event version\",\n          \"type\": \"string\"\n        }\n      },\n      \"required\": [\n        \"name\",\n        \"data\",\n        \"user\"\n      ],\n      \"type\": \"object\"\n    }\n  }\n]"

