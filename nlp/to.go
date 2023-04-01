package nlp

import "strings"

// Converts s to the same case as like.
// Handles upper, lower, and title case.
func ToSameCase(s string, like string) string {
	if IsUpper(like) {
		return strings.ToUpper(s)
	} else if IsLower(like) {
		return strings.ToLower(s)
	} else if IsTitle(like) {
		return strings.Title(s)
	} else {
		return s
	}
}
