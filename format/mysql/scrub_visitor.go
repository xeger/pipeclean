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
	switch stmt.(type) {
	case *ast.InsertStmt:
		if doInserts {
			v.insert = &insertState{}
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
	switch st := in.(type) {
	case *ast.TableName:
		if v.insert != nil {
			v.insert.tableName = st.Name.L
		}
	case *ast.ColumnName:
		// insert column names present in SQL source; accumulate them
		if v.insert != nil {
			v.insert.columnNames = append(v.insert.columnNames, st.Name.L)
		}
	case *test_driver.ValueExpr:
		if v.insert != nil {
			// column names omitted from SQL source; infer from table schema
			if v.insert.valueIndex == 0 && len(v.insert.columnNames) == 0 {
				v.insert.columnNames = v.ctx.TableColumns[v.insert.tableName]
			}
			defer func() {
				v.insert.valueIndex++
			}()
			switch st.Kind() {
			case test_driver.KindString:
				datum := test_driver.Datum{}
				s := st.Datum.GetString()
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
