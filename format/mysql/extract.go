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

	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		values := extract(v, p, line)
		for _, v := range values {
			w.Write([]byte(v + "\n"))
		}
	}
}
