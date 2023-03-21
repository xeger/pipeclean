package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

func parse(sql string) (ast.StmtNode, error) {
	p := parser.New()

	stmtNodes, _, err := p.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}

	if len(stmtNodes) != 1 {
		return nil, fmt.Errorf("invalid statement")
	}

	return stmtNodes[0], nil
}

func sanitize(stmt ast.StmtNode) ast.StmtNode {
	switch stmt.(type) {
	case *ast.UpdateStmt:
		// TODO: sanitize update statement
		return stmt
	default:
		return stmt
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		stmt, _ := parse(line)
		if stmt != nil {
			fmt.Println(stmt.Text())
		} else {
			fmt.Print(line)
		}
	}
}
