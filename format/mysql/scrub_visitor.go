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
func (sv *scrubVisitor) ScrubStatement(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch stmt.(type) {
	case *ast.InsertStmt:
		if doInserts {
			sv.insert = &insertState{}
			stmt.Accept(sv)
			sv.insert = nil
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

func (sv *scrubVisitor) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *ast.TableName:
		if sv.insert != nil {
			sv.insert.tableName = st.Name.L
		}
	case *ast.ColumnName:
		// insert column names present in SQL source; accumulate them
		if sv.insert != nil {
			sv.insert.columnNames = append(sv.insert.columnNames, st.Name.L)
		}
	case *test_driver.ValueExpr:
		if sv.insert != nil {
			// column names omitted from SQL source; infer from table schema
			if sv.insert.valueIndex == 0 && len(sv.insert.columnNames) == 0 {
				sv.insert.columnNames = sv.ctx.TableColumns[sv.insert.tableName]
			}
			defer func() {
				sv.insert.valueIndex++
			}()
			switch st.Kind() {
			case test_driver.KindString:
				datum := test_driver.Datum{}
				s := st.Datum.GetString()
				if sv.scrubber.EraseString(s, sv.insert.ColumnName()) {
					datum.SetNull()
				} else {
					datum.SetString(sv.scrubber.ScrubString(s, sv.insert.ColumnName()))
				}
				return &test_driver.ValueExpr{Datum: datum}, true
			}
		}
	}
	return in, false
}

func (sc *scrubVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
