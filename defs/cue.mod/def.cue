package eventdefintions

#Def: {
	// A short description of the event.
	description?: string

	schema: {
		// The unique name of the event
		name: string
		// The event payload, containing all event data
		data: [string]: _
		// User information for the author of the event
		user: [string]: _

		// An optional event version
		v?: string
	}

	examples?: [...string]
}
