package scrubbing

import (
	"bytes"
	"fmt"

	"github.com/pingcap/tidb/parser"
	_ "github.com/pingcap/tidb/parser/test_driver"
	"github.com/xeger/sqlstream/nlp"
)

// Scrub a sequence of lines received via in. Each line may comprise multiple statements,
// which will recombined with a newline separator and transmitted to out.
//
// Because of the 1:1 mapping between sends and receives, this function can be used with
// buffered channels provided the caller takes care to preserve ordering.
func Scrub(models []nlp.Model, confidence float64, in <-chan string, out chan<- string) {
	p := parser.New()
	sc := NewScrubber(models, confidence)
	for line := range in {
		buf := bytes.NewBufferString("")

		stmts, _, err := p.Parse(line, "", "")
		if (err != nil || len(stmts) == 0) && doComments {
			fmt.Fprint(buf, line)
		}

		for _, in := range stmts {
			out, processed := sc.Scrub(in)
			if !processed {
				fmt.Fprintln(buf, out.OriginalText())
			} else if out != nil {
				fmt.Fprintln(buf, restore(out))
			}
		}

		out <- buf.String()
	}
}
