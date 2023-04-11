package mysql

import (
	"bytes"
	"fmt"

	"github.com/pingcap/tidb/parser"
	"github.com/xeger/pipeclean/scrubbing"
)

func scrub(msc *mysqlScrubber, p *parser.Parser, line string) string {
	buf := bytes.NewBufferString("")

	stmts, _, err := p.Parse(line, "", "")
	if (err != nil || len(stmts) == 0) && doComments {
		fmt.Fprint(buf, line)
	}

	for _, in := range stmts {
		out, processed := msc.ScrubStatement(in)
		if !processed {
			fmt.Fprintln(buf, out.OriginalText())
		} else if out != nil {
			fmt.Fprintln(buf, restore(out))
		}
	}

	return buf.String()
}

// Scrub sanitizes a single line, which may contain multiple SQL statements.
func Scrub(sc *scrubbing.Scrubber, line string) string {
	msc := &mysqlScrubber{sc}
	p := parser.New()
	return scrub(msc, p, line)
}

// ScrubChan sanitizes a sequence of lines, each of which may contain multiple
// SQL statements. It sends one string for every string received, allowing the
// caller to handle parallelism.
func ScrubChan(sc *scrubbing.Scrubber, in <-chan string, out chan<- string) {
	msc := &mysqlScrubber{sc}
	p := parser.New()
	for line := range in {
		out <- scrub(msc, p, line)
	}
}
