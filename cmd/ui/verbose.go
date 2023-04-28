package ui

import (
	"fmt"
	"os"
	"strings"
)

var IsVerbose = false

func Verbose(err error) Hinter {
	return Verbosef(err.Error())
}

func Verbosef(format string, args ...interface{}) Hinter {
	if strings.Index(format, "\n") <= 0 {
		format = format + "\n"
	}
	if IsVerbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
	return &hinter{suppress: !IsVerbose}
}
