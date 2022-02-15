Status: "open" | "closed"

#Some: {
	with: string
	...
}

#Event: {
	name: string
	data: {
		action:          "push" | "pull" | "rebase"
		status:          Status
		number:          uint & <=10
		static:          "lol this is content"
		optionalStatic?: "some opt content"
		staticNumber:    1
		staticBool?:     true
		enabled:         bool
		numeric:         number
		mixed:           string | int
		friends: [...{
			id:   int
			name: string
		}]
		nested: [...{
			id:   int
			heyy: "what" | "do"
		}]
	}
	allow: #Some & {
		included: bool
	}
	numberList: [...(int | float)]
	fixedNumber: [1, 2, 3.14159]
}
