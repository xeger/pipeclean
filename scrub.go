package main

import (
	"bytes"
	"fmt"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

// Preserves non-parseable lines (assuming they are comments).
const doComments = true

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = true

// Attempts to remove sensitive data from an AST. Returns nil if the entire statement should be dropped.
func scrubStmt(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch st := stmt.(type) {
	// for table name: st.Table.TableRefs.Left.(*ast.TableSource).Source.(*ast.TableName).Name
	// for raw values: st.Lists[0][0], etc...
	case *ast.InsertStmt:
		if doInserts {
			st.Accept(NewScrubber())
			return st, true
		} else {
			return nil, true
		}
	default:
		if doMisc {
			return stmt, false
		}
		return nil, true
	}
}

// Parallelizes scrubbing.
func scrubLines(in <-chan string, out chan<- string) {
	p := parser.New()
	for line := range in {
		buf := bytes.NewBufferString("")

		stmts, _, err := p.Parse(line, "", "")
		if (err != nil || len(stmts) == 0) && doComments {
			fmt.Fprint(buf, line)
		}

		for _, in := range stmts {
			out, processed := scrubStmt(in)
			if !processed {
				fmt.Fprintln(buf, out.OriginalText())
			} else if out != nil {
				fmt.Fprintln(buf, restore(out))
			}
		}

		out <- buf.String()
	}
}
