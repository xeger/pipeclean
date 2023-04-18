package mysql

import (
	"github.com/pingcap/tidb/parser"
)

// ScrubContext accumulates information about the structure of input data
// which can later be used to scrub the same data.
type ScrubContext struct {
	TableColumns map[string][]string
}

func (sc *ScrubContext) Scan(sql string) error {
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

func NewScrubContext() *ScrubContext {
	return &ScrubContext{
		TableColumns: make(map[string][]string),
	}
}
