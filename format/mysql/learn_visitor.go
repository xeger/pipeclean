package mysql

import (
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
)

type learnVisitor struct {
	ctx    *Context
	insert *insertState
	models map[string]nlp.Model
	policy *scrubbing.Policy
}

// LearnStatement trains models based on values in a SQL insert AST.
func (v *learnVisitor) LearnStatement(stmt ast.StmtNode) {
	switch typed := stmt.(type) {
	case *ast.InsertStmt:
		v.insert = newInsertState(typed)
		stmt.Accept(v)
		v.insert = nil
	}
}

func (v *learnVisitor) Enter(in ast.Node) (ast.Node, bool) {
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
				disposition, _ := v.policy.MatchFieldName(v.insert.Names())
				switch disposition.Action() {
				case "generate":
					model := v.models[disposition.Parameter()]
					if model != nil {
						model.Train(typed.Datum.GetString())
					}
				}
				return typed, true
			}
		}
	}
	return in, false
}

func (v *learnVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
