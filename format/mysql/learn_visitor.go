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
func (lv *learnVisitor) LearnStatement(stmt ast.StmtNode) {
	switch stmt.(type) {
	case *ast.InsertStmt:
		lv.insert = &insertState{}
		stmt.Accept(lv)
		lv.insert = nil
	}
}

func (lv *learnVisitor) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *ast.TableName:
		if lv.insert != nil {
			lv.insert.tableName = st.Name.L
		}
	case *ast.ColumnName:
		// insert column names present in SQL source; accumulate them
		if lv.insert != nil {
			lv.insert.columnNames = append(lv.insert.columnNames, st.Name.L)
		}
	case *test_driver.ValueExpr:
		if lv.insert != nil {
			// column names omitted from SQL source; infer from table schema
			if lv.insert.valueIndex == 0 && len(lv.insert.columnNames) == 0 {
				lv.insert.columnNames = lv.ctx.TableColumns[lv.insert.tableName]
			}
			defer func() {
				lv.insert.valueIndex++
			}()
			switch st.Kind() {
			case test_driver.KindString:
				datum := test_driver.Datum{}
				d := lv.policy.MatchFieldName(lv.insert.ColumnName())
				switch d.Action() {
				case "generate":
					model := lv.models[d.Parameter()]
					if model != nil {
						model.Train(datum.GetString())
					}
				}
				return &test_driver.ValueExpr{Datum: datum}, true
			}
		}
	}
	return in, false
}

func (lv *learnVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
