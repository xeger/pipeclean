package mysql

import (
	"context"

	"github.com/pingcap/tidb/parser"
)

// Context accumulates information about the structure of input data
// which can later be used to scrub the same data.
type Context struct {
	context.Context
	TableColumns map[string][]string
}

func (sc *Context) Scan(sql string) error {
	p := parser.New()
	stmts, _, err := p.Parse(sql, "", "")
	if err != nil {
		return err
	}
	siv := &schemaInfoVisitor{info: sc}
	for _, in := range stmts {
		siv.ScanStatement(in)
	}
	return nil
}

func NewContext() *Context {
	return &Context{
		Context:      context.Background(),
		TableColumns: make(map[string][]string),
	}
}
