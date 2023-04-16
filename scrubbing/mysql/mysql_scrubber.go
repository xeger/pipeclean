package mysql

import (
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
	"github.com/xeger/pipeclean/scrubbing"
)

type insertState struct {
	// Name of the table being inserted into.
	tableName string
	// List of column names (explicitly specified in current statement, or inferred from table schema).
	columnNames []string
	// Number of ValueExpr seen so far across all rows of current statement.
	valueIndex int
}

// ColumnName infers the name of the column to which the next ValueExpr will apply.
// It returns the empty string if the column name is unknown.
func (is insertState) ColumnName() string {
	if len(is.columnNames) == 0 {
		return ""
	}
	return is.columnNames[is.valueIndex%len(is.columnNames)]
}

type mysqlScrubber struct {
	*scrubbing.Scrubber
	insert *insertState
}

// ScrubStatement sensitive data from an SQL AST.
// May modify the AST in-place (and return it), or may return a derived AST.
// Returns nil if the entire statement should be omitted from output.
func (sc *mysqlScrubber) ScrubStatement(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch st := stmt.(type) {
	case *ast.InsertStmt:
		if doInserts {
			sc.insert = &insertState{}
			st.Accept(sc)
			sc.insert = nil
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
	case *ast.TableName:
		if sc.insert != nil {
			sc.insert.tableName = st.Name.L
		}
	case *ast.ColumnName:
		if sc.insert != nil {
			sc.insert.columnNames = append(sc.insert.columnNames, st.Name.L)
		}
	case *test_driver.ValueExpr:
		if sc.insert != nil {
			if sc.insert.valueIndex == 0 && len(sc.insert.columnNames) == 0 {
				// TODO: grab column names from schema definition
			}
			defer func() {
				sc.insert.valueIndex++
			}()
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
	}
	return in, false
}

func (sc *mysqlScrubber) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
