package eventdefintions

#Def: {
	// integration defines the name of the integration that generates this
	// event.
	integration?: string

	schema: {
		// The unique name of the event
		name: string
		// The event payload, containing all event data
		data: [string]: _
		// User information for the author of the event
		user: [string]: _
	}

	example?: string
}
