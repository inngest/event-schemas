package eventdefintions

#Def: {
	// integration defines the name of the integration that generates this
	// event.
	integration?: string

	identifiers?: {
		[string]: string
	}

	schema: {
		name: string
		data: [string]: _
	}
}
