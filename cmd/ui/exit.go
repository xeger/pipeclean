package ui

import "os"

type Reason rune

const (
	AssertionFailed  = Reason('!')
	InvalidArgs      = Reason('-')
	InvalidInputFile = Reason('>')
	ToDo             = Reason(':')
)

func Exit(reason Reason) {
	os.Exit(int(reason))
}

func ExitBug(issue string) {
	Fatalf(issue)
	Exit(AssertionFailed)
}

func ExitNotImplemented(feature string) {
	Fatalf("Not yet implemented: %s", feature)
	Exit(ToDo)
}
