# Event schemas

This repository contains event schema definitions for common events ingested within
[Inngest](https://www.inngest.com).  This allows you to get full type information for common events,
and allows you to generate schemas in multiple languages for each event.

Events are defined using [cue](https://cuelang.org), a novel declarative data language.  It's
concise, strongly typed, allows constraining values, defaults, and annotations for extra data.

## Converting JSON Schema & OpenAPI definitions to Cue

We use [cue](https://cuelang.org) as our canonical representation of event types.  You can generate
a Cue type definition from an existing JSON schema or API schema:

```
cue import jsonschema ./path/to/schema.json -o -
```

This will print the cue type definitions to stdout.  You can then take these definitions and add
them to ./defs/${service.cue} to document events.
