package ui

import (
	"fmt"
	"os"
	"strings"
)

func Warn(err error) Hinter {
	return Warnf(err.Error())
}

func Warnf(format string, args ...interface{}) Hinter {
	if strings.Index(format, "\n") <= 0 {
		format = format + "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	return &hinter{}
}
