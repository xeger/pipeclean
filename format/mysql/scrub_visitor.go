package mysql

import (
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
	"github.com/xeger/pipeclean/scrubbing"
)

type scrubVisitor struct {
	ctx      *Context
	scrubber *scrubbing.Scrubber
	insert   *insertState
}

// ScrubStatement sensitive data from an SQL AST.
// May modify the AST in-place (and return it), or may return a derived AST.
// Returns nil if the entire statement should be omitted from output.
func (v *scrubVisitor) ScrubStatement(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch typed := stmt.(type) {
	case *ast.InsertStmt:
		if doInserts {
			v.insert = newInsertState(typed)
			stmt.Accept(v)
			v.insert = nil
			return stmt, true
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

func (v *scrubVisitor) Enter(in ast.Node) (ast.Node, bool) {
	switch typed := in.(type) {
	case *ast.TableName:
		if v.insert != nil {
			v.insert.tableName = typed.Name.L
		}
	case *ast.ColumnName:
		// insert column names present in SQL source; accumulate them
		if v.insert != nil {
			v.insert.columnNames = append(v.insert.columnNames, typed.Name.L)
		}
	case *test_driver.ValueExpr:
		if v.insert != nil {
			v.insert.ObserveContext(v.ctx)
			defer func() {
				v.insert.Advance()
			}()
			switch typed.Kind() {
			case test_driver.KindString:
				datum := test_driver.Datum{}
				s := typed.Datum.GetString()
				names := v.insert.Names()
				if v.scrubber.EraseString(s, names) {
					datum.SetNull()
				} else {
					datum.SetString(v.scrubber.ScrubString(s, names))
				}
				return &test_driver.ValueExpr{Datum: datum}, true
			}
		}
	}
	return in, false
}

func (v *scrubVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
