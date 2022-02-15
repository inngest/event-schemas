{
	array: [...{
		id: string
		ok: bool
	}]
	mixed_array: [...{
		id:     string
		number: float
		ok:     bool
	} | {
		id:     string
		number: int
		ok:     bool
	} | {
		id:          string
		number:      float
		another_str: string
	}]
}
