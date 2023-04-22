package json

import (
	"encoding/json"
	"io"

	"github.com/xeger/pipeclean/scrubbing"
)

// Scrub sanitizes a JSON document, which may be part of a stream of documents
// a sub-object of some larger document.
//
// The caller is responsible for parsing streams into sub-objects and must pass
// only well-formed, complete documents to Scrub.
func Scrub(sc *scrubbing.Scrubber, r io.Reader, w io.Writer) {
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)
	var v any
	for err := dec.Decode(&v); err == nil; err = dec.Decode(&v) {
		sc.ScrubData(v, nil)
		enc.Encode(v)
	}
}
