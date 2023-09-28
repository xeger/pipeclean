package mysql

import (
	"fmt"

	"github.com/pingcap/tidb/parser/ast"
)

type insertState struct {
	// Name of the table being inserted into.
	tableName string
	// List of column names (explicitly specified in current statement, or inferred from table schema).
	columnNames []string
	// Value tuple sizes of the current statement.
	rowLength int
	// Number of ValueExpr seen so far across all rows of current statement.
	valueIndex int
}

func newInsertState(stmt *ast.InsertStmt) *insertState {
	rowLength := 0
	for _, list := range stmt.Lists {
		if rowLength == 0 {
			rowLength = len(list)
		} else if len(list) != rowLength {
			// TODO: handle this case by storing an array of row-tuple lengths & iterating through it
			panic(fmt.Sprintf("inconsistent INSERT row lengths: %d prior vs %d next", rowLength, len(list)))
		}
	}
	return &insertState{rowLength: rowLength}
}

// Advance increments the column-value index so that Names() remains accurate.
func (is *insertState) Advance() {
	is.valueIndex += 1
}

// Names returns a list of column names to which the next ValueExpr will apply.
// The list contains 0-3 elements depending on the completeness of the schema
// information provided in context.
func (is *insertState) Names() []string {
	colIdx := is.valueIndex
	if is.rowLength > 0 {
		colIdx = colIdx % is.rowLength
	}
	names := make([]string, 0, 3)
	if len(is.tableName) > 0 {
		if len(is.columnNames) > 0 {
			colName := is.columnNames[colIdx]
			names = append(names, colName)
			names = append(names, fmt.Sprintf("%s.%s", is.tableName, colName))
		}
		names = append(names, fmt.Sprintf("%s.%d", is.tableName, colIdx))
	}

	return names
}

// If column names were omitted from the SQL INSERT statement, infer them from the previously-scanned table schema.
func (is *insertState) ObserveContext(ctx *Context) {
	if is.valueIndex == 0 && len(is.columnNames) == 0 {
		is.columnNames = ctx.TableColumns[is.tableName]
	}
}
