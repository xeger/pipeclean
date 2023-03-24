package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
	_ "github.com/pingcap/tidb/parser/test_driver"

	"gonum.org/v1/gonum/mathext/prng"
)

// Preserves non-parseable lines (assuming they are comments).
const doComments = true

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = true

var flagCPU = flag.Int("c", runtime.NumCPU(), "parallelism (default: number of CPUs)")

// Turns an AST back into a string.
func restore(stmt ast.StmtNode) string {
	buf := new(bytes.Buffer)
	ctx := format.NewRestoreCtx(format.RestoreKeyWordUppercase|format.RestoreNameBackQuotes|format.RestoreStringSingleQuotes|format.RestoreStringWithoutDefaultCharset, buf)
	err := stmt.Restore(ctx)
	if err != nil {
		panic(err)
	}
	s := buf.String()
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s + ";\n"
}

// Attempts to remove sensitive data from an AST. Returns nil if the entire statement should be dropped.
func scrubStmt(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch st := stmt.(type) {
	// for table name: st.Table.TableRefs.Left.(*ast.TableSource).Source.(*ast.TableName).Name
	// for raw values: st.Lists[0][0], etc...
	case *ast.InsertStmt:
		if doInserts {
			v := &scrubber{source: prng.NewMT19937()}
			st.Accept(v)
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

func main() {
	flag.Parse()
	N := *flagCPU

	in := make([]chan string, N)
	out := make([]chan string, N)
	for i := 0; i < N; i++ {
		in[i] = make(chan string)
		out[i] = make(chan string)
		go scrubLines(in[i], out[i])
	}
	drain := func(to int) {
		for i := 0; i < to; i++ {
			fmt.Print(<-out[i])
		}
	}
	done := func() {
		for i := 0; i < N; i++ {
			close(in[i])
			close(out[i])
		}
	}

	reader := bufio.NewReader(os.Stdin)
	l := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		in[l] <- line
		l = (l + 1) % N
		if l == 0 {
			drain(N)
		}
	}
	drain(l)
	done()
}
