package mysql

import (
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
)

type extractVisitor struct {
	ctx    *Context
	names  []string
	insert *insertState
	values []string
}

// ExtractStatement pulls interesting field values from INSERT statements.
func (v *extractVisitor) ExtractStatement(stmt ast.StmtNode) []string {
	switch typed := stmt.(type) {
	case *ast.InsertStmt:
		v.insert = newInsertState(typed)
		v.values = []string{}
		stmt.Accept(v)
		v.insert = nil
	}
	return v.values
}

func (v *extractVisitor) Enter(in ast.Node) (ast.Node, bool) {
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
				if v.MatchFieldName(v.insert.Names()) {
					v.values = append(v.values, typed.Datum.GetString())
				}
				return typed, true
			}
		}
	}
	return in, false
}

func (v *extractVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (v *extractVisitor) MatchFieldName(names []string) bool {
	for _, want := range v.names {
		for _, got := range names {
			if want == got {
				return true
			}
		}
	}
	return false
}
