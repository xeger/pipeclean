package ui

import (
	"fmt"
	"os"
)

type Hinter interface {
	Hint(hints ...string) Hinter
}

type hinter struct{}

func (h *hinter) Hint(hints ...string) Hinter {
	for _, hint := range hints {
		fmt.Fprintln(os.Stderr, "  "+hint)
	}
	return h
}
