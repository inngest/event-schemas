Status: "open" | "closed"

#Event: {
	name: string
	data: {
		action:          "push" | "pull"
		status:          Status
		number:          uint & >0
		static:          "lol this is content"
		optionalStatic?: "some opt content"
		staticNumber:    1
		staticBool?:     true
		enabled:         bool
		numeric:         number
	}
	numberList: [...(int | float)]
	fixedNumber: [1, 2, 3.14159]
}
