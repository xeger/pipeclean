package cmd

// Shared flag values for all commands.
// Helps ensure consistency i.e. if "foo" is a float64 in A, then it should be
// a float64 in B, too.
var (
	confidence  float64
	format      string
	parallelism int
	salt        string
)
