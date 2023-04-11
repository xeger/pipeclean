package mysql

import (
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
	"github.com/xeger/pipeclean/scrubbing"
)

type mysqlScrubber struct {
	*scrubbing.Scrubber
}

// Preserves non-parseable lines (assuming they are comments).
const doComments = true

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = true

// Removes sensitive data from an SQL statement AST.
// May modify the AST in-place (and return it), or may return a derived AST.
// Returns nil if the entire statement should be dropped.
func (sc *mysqlScrubber) ScrubStatement(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch st := stmt.(type) {
	// for table name: st.Table.TableRefs.Left.(*ast.TableSource).Source.(*ast.TableName).Name
	// for raw values: st.Lists[0][0], etc...
	case *ast.InsertStmt:
		if doInserts {
			st.Accept(sc)
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

func (sc *mysqlScrubber) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *test_driver.ValueExpr:
		switch st.Kind() {
		case test_driver.KindString:
			datum := test_driver.Datum{}
			s := st.Datum.GetString()
			if sc.EraseString(s) {
				datum.SetNull()
			} else {
				datum.SetString(sc.ScrubString(s))
			}
			return &test_driver.ValueExpr{Datum: datum}, true
		}
	}
	return in, false
}

func (sc *mysqlScrubber) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
