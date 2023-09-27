package mysql

import (
	"github.com/pingcap/tidb/parser/ast"
)

type schemaInfoVisitor struct {
	info      *Context
	columnDef bool
	tableName string
}

func (v *schemaInfoVisitor) ScanStatement(stmt ast.StmtNode) {
	switch stmt.(type) {
	case *ast.CreateTableStmt:
		v.tableName = ""
		stmt.Accept(v)
	}
}

func (v *schemaInfoVisitor) Enter(in ast.Node) (ast.Node, bool) {
	switch typed := in.(type) {
	case *ast.TableName:
		v.tableName = typed.Name.L
		if v.info.TableColumns[v.tableName] == nil {
			v.info.TableColumns[v.tableName] = make([]string, 0, 32)
		}
	case *ast.ColumnDef:
		v.columnDef = true
	case *ast.ColumnName:
		if v.columnDef {
			v.info.TableColumns[v.tableName] = append(v.info.TableColumns[v.tableName], typed.Name.L)
		}
	}
	return in, false
}

func (v *schemaInfoVisitor) Leave(in ast.Node) (ast.Node, bool) {
	switch in.(type) {
	case *ast.ColumnDef:
		v.columnDef = false
	}
	return in, true
}
