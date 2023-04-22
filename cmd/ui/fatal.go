package ui

import (
	"fmt"
	"os"
	"strings"
)

func Fatal(err error) Hinter {
	return Fatalf(err.Error())
}

func Fatalf(format string, args ...interface{}) Hinter {
	if strings.Index(format, "\n") <= 0 {
		format = format + "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	return &hinter{}
}
