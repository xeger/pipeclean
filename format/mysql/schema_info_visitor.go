package mysql

import (
	"github.com/pingcap/tidb/parser/ast"
)

type schemaInfoVisitor struct {
	info      *Context
	columnDef bool
	tableName string
}

func (siv *schemaInfoVisitor) ScanStatement(stmt ast.StmtNode) {
	switch stmt.(type) {
	case *ast.CreateTableStmt:
		siv.tableName = ""
		stmt.Accept(siv)
	}
}

func (siv *schemaInfoVisitor) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *ast.TableName:
		siv.tableName = st.Name.L
		if siv.info.TableColumns[siv.tableName] == nil {
			siv.info.TableColumns[siv.tableName] = make([]string, 0, 32)
		}
	case *ast.ColumnDef:
		siv.columnDef = true
	case *ast.ColumnName:
		if siv.columnDef {
			siv.info.TableColumns[siv.tableName] = append(siv.info.TableColumns[siv.tableName], st.Name.L)
		}
	}
	return in, false
}

func (siv *schemaInfoVisitor) Leave(in ast.Node) (ast.Node, bool) {
	switch in.(type) {
	case *ast.ColumnDef:
		siv.columnDef = false
	}
	return in, true
}
