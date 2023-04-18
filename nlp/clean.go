package nlp

import "unicode"

// Clean returns the lower-case representation of input with
// whitespace trimmed and normalized to a single 0x20 (SPC) character.
func Clean(input string) string {
	output := make([]rune, 0, len(input))

	for _, c := range input {
		if unicode.IsLower(c) {
			output = append(output, c)
		} else if unicode.IsUpper(c) {
			output = append(output, unicode.ToLower(c))
		} else if unicode.IsSpace(c) {
			if len(output) > 0 && !unicode.IsSpace(output[len(output)-1]) {
				output = append(output, ' ')
			}
		} else {
			output = append(output, c)
		}
	}

	if len(output) > 0 && unicode.IsSpace(output[len(output)-1]) {
		output = output[:len(output)-1]
	}

	return string(output)
}

// CleanToken calls Clean on input, then removes all non-alphanumeric
// characters from it. The result is useful for seeding a PRNG in a
// way that disregards punctuation and case.
func CleanToken(input string) string {
	input = Clean(input)
	output := make([]rune, 0, len(input))
	for _, c := range input {
		if unicode.IsLower(c) || (c >= '0' && c <= '9') {
			output = append(output, c)
		}
	}
	return string(output)
}
