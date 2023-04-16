package mysql

import (
	"bytes"
	"fmt"

	"github.com/pingcap/tidb/parser"
	"github.com/xeger/pipeclean/scrubbing"
)

func scrub(sv *scrubVisitor, p *parser.Parser, line string) string {
	buf := bytes.NewBufferString("")

	stmts, _, err := p.Parse(line, "", "")
	if (err != nil || len(stmts) == 0) && doComments {
		fmt.Fprint(buf, line)
	}

	for _, in := range stmts {
		out, processed := sv.ScrubStatement(in)
		if !processed {
			fmt.Fprintln(buf, out.OriginalText())
		} else if out != nil {
			fmt.Fprintln(buf, restore(out))
		}
	}

	return buf.String()
}

// ScrubChan sanitizes a sequence of lines, each of which may contain multiple
// SQL statements. It sends one output string for every input string received,
// even for multi-line or multi-statement inputs. This allows the caller to
// handle parallelism as desired.
func ScrubChan(ctx *ScrubContext, sc *scrubbing.Scrubber, in <-chan string, out chan<- string) {
	sv := &scrubVisitor{ctx, sc, nil}
	p := parser.New()
	for line := range in {
		out <- scrub(sv, p, line)
	}
}
