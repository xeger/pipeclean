package nlp

import "fmt"

var ErrInvalidModel = fmt.Errorf(`Mismatch between model configuration and persisted model.`)
