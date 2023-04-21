package nlp

import (
	"path/filepath"
	"strings"
)

type Model interface {
	Recognize(input string) float64
	Train(input string)
}

// MfBase returns the file's base name with no extensions.
func mfBase(filename string) string {
	base := filepath.Base(filename)
	dot := strings.Index(base, ".")
	if dot < 0 {
		return base
	}
	return base[:dot]
}

// MfExt returns the last two extensions of the file's base name.
func mfExt(filename string) string {
	base := filepath.Base(filename)
	suffix2 := filepath.Ext(base)
	suffix1 := filepath.Ext(strings.TrimSuffix(base, suffix2))
	return strings.Join([]string{suffix1, suffix2}, "")
}
