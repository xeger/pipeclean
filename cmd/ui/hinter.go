package ui

import (
	"fmt"
	"os"
)

type hinter struct{}

func (h hinter) Hint(hints ...string) hinter {
	for _, hint := range hints {
		fmt.Fprintln(os.Stderr, "  "+hint)
	}
	return h
}
