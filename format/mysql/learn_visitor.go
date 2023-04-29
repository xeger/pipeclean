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
	switch stmt.(type) {
	case *ast.InsertStmt:
		v.insert = &insertState{}
		stmt.Accept(v)
		v.insert = nil
	}
}

func (v *learnVisitor) Enter(in ast.Node) (ast.Node, bool) {
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
				disposition, _ := v.policy.MatchFieldName(v.insert.Names())
				switch disposition.Action() {
				case "generate":
					model := v.models[disposition.Parameter()]
					if model != nil {
						model.Train(st.Datum.GetString())
					}
				}
				return st, true
			}
		}
	}
	return in, false
}

func (v *learnVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
