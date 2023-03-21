package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
	"github.com/pingcap/tidb/parser/test_driver"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

// Preserves non-parseable lines (assuming they are comments).
const COMMENTS = true

// Preserves INSERT statements (disable to make debug printfs readable).
const INSERTS = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const MISC = true

// Does silly trivial scrambling as POC
const REVERSE = false

type scrubber struct {
}

func (v *scrubber) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *test_driver.ValueExpr:
		switch st.Kind() {
		case test_driver.KindString:
			scrubbed := test_driver.Datum{}
			scrubbed.SetString(v.scrubString(st.Datum.GetString()))
			return &test_driver.ValueExpr{Datum: scrubbed}, true
		}
	}
	return in, false
}

func (v *scrubber) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (v *scrubber) scrubString(s string) string {
	// TODO: recognize (& sanitize?) all well-formed YAML, JSON
	if strings.Index(s, "\n") >= 0 {
		return s
	}
	if REVERSE {
		result := ""
		for _, v := range s {
			result = string(v) + result
		}
		return result
	}
	return s
}

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
	switch st := stmt.(type) {
	case *ast.InsertStmt:
		v := &scrubber{}
		// fmt.Printf("%+v\n", st.Table.TableRefs.Left.(*ast.TableSource).Source.(*ast.TableName).Name)
		// fmt.Printf("%d %+v\n", len(st.Lists), st.Lists[0][0])
		st.Accept(v)
		buf := new(bytes.Buffer)
		ctx := format.NewRestoreCtx(format.DefaultRestoreFlags, buf)
		st.Restore(ctx)
		s := buf.String()
		s = strings.ReplaceAll(s, "\n", "\\n")
		s = s + ";"
		if INSERTS {
			st, _ := parse(s)
			return st
		} else {
			return nil
		}
	default:
		if MISC {
			return stmt
		}
		return nil
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
			stmt = sanitize(stmt)
			if stmt != nil {
				fmt.Println(stmt.Text())
			}
		} else {
			if COMMENTS {
				fmt.Print(line)
			}
		}
	}
}
