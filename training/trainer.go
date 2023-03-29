package main

import (
	"fmt"
	"os"

	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

type trainer struct {
}

func (tr *trainer) Enter(in ast.Node) (ast.Node, bool) {
	fmt.Fprintf(os.Stderr, "Enter>%t\n", in)
	return in, false
}

func (tr *trainer) Leave(in ast.Node) (ast.Node, bool) {
	fmt.Fprintf(os.Stderr, "Leave>%t\n", in)
	return in, false
}
