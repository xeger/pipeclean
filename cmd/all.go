package cmd

// Shared flag values for all commands.
//
// Helps ensure consistency i.e. if "foo" is a float64 in command A, then
// it must be a float64 in command B, too.
var (
	confidence  float64
	context     []string
	mode        string
	parallelism int
	policy      string
	salt        string
)
