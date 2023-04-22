package ui

import (
	"fmt"
	"os"
	"strings"
)

func Fatalf(format string, args ...interface{}) hinter {
	if strings.Index(format, "\n") <= 0 {
		format = format + "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	return hinter{}
}
