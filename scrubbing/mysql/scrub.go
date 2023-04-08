package mysql

import (
	"bytes"
	"fmt"

	"github.com/pingcap/tidb/parser"
	"github.com/xeger/sqlstream/scrubbing"
)

// Preserves non-parseable lines (assuming they are comments).
const doComments = false

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = false

func Scrub(sc *scrubbing.Scrubber, in <-chan string, out chan<- string) {
	msc := &mysqlScrubber{sc}

	p := parser.New()
	for line := range in {
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

		out <- buf.String()
	}
}
