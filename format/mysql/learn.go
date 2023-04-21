package mysql

import (
	"github.com/pingcap/tidb/parser"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
)

func learn(lv *learnVisitor, p *parser.Parser, line string) {
	stmts, _, _ := p.Parse(line, "", "")
	for _, in := range stmts {
		lv.LearnStatement(in)
	}
}

// TODO docs
func LearnChan(ctx *Context, models map[string]nlp.Model, policy *scrubbing.Policy, in <-chan string) {
	lv := &learnVisitor{ctx, nil, models, policy}
	p := parser.New()
	for line := range in {
		learn(lv, p, line)
	}
}
