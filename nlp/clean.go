package nlp

import "unicode"

// Clean returns the lower-case representation of the input string with
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
