package mysql

import (
	"context"
	"fmt"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
)

type learnVisitor struct {
	ctx context.Context
}

func (lv *learnVisitor) LearnStatement(stmt ast.StmtNode) (out string, processed bool) {
	panic("TODO")
}

func learn(lv any, p *parser.Parser, line string) {
	stmts, _, err := p.Parse(line, "", "")
	if err != nil || len(stmts) == 0 {
		return
	}

	for _, in := range stmts {
		// TODO learn
		fmt.Println(in)
	}
}

func LearnChan(ctx context.Context, in <-chan string) {
	lv := &learnVisitor{ctx}
	p := parser.New()
	for line := range in {
		learn(lv, p, line)
	}
}
