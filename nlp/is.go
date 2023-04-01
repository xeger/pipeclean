package nlp

import "unicode"

// Determines whether all alphabetic characters of s are lower case.
func IsLower(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return false
		}
	}

	return true
}

// Determines whether all words of s are title case.
func IsTitle(s string) bool {
	var last rune
	for _, r := range s {
		if unicode.IsLower(r) && (last == ' ' || last == 0) {
			return false
		} else if unicode.IsUpper(r) && (last != ' ' && last != 0) {
			return false
		}
		last = r
	}

	return true
}

// Determines whether all alphabetic characters of s are upper case.
func IsUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return false
		}
	}

	return true
}
