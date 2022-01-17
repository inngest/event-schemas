package parse

import "cuelang.org/go/cue"

func cueField(v cue.Value, field string) cue.FieldInfo {
	f, _ := v.LookupField(field)
	return f
}

func cueString(v cue.Value, field string) string {
	str, _ := cueField(v, field).Value.String()
	return str
}
