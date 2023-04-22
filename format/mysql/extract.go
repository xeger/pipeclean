package mysql

import (
	"bufio"
	"io"

	"github.com/pingcap/tidb/parser"
)

func extract(v *extractVisitor, p *parser.Parser, line string) []string {
	stmts, _, _ := p.Parse(line, "", "")
	values := []string{}

	for _, in := range stmts {
		out := v.ExtractStatement(in)
		values = append(values, out...)
	}

	return values
}

func Extract(ctx *Context, names []string, r io.Reader, w io.Writer) {
	p := parser.New()
	v := &extractVisitor{ctx, names, nil, nil}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		values := extract(v, p, scanner.Text())
		for _, v := range values {
			w.Write([]byte(v + "\n"))
		}
	}
}
