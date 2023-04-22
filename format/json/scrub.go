package json

import (
	"encoding/json"
	"io"

	"github.com/xeger/pipeclean/scrubbing"
)

// Scrub sanitizes a single line, which may contain multiple SQL statements.
func Scrub(sc *scrubbing.Scrubber, r io.Reader, w io.Writer) {
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)
	var v any
	for err := dec.Decode(&v); err == nil; err = dec.Decode(&v) {
		sc.ScrubData(v, nil)
		enc.Encode(v)
	}
}
