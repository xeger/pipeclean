package mysql

import "fmt"

type insertState struct {
	// Name of the table being inserted into.
	tableName string
	// List of column names (explicitly specified in current statement, or inferred from table schema).
	columnNames []string
	// Number of ValueExpr seen so far across all rows of current statement.
	valueIndex int
}

// Names returns a list of column names to which the Next ValueExpr will apply.
// The list contains 0-3 elements depending on the completeness of the schema
// information provided in context.
func (is insertState) Names() []string {
	names := make([]string, 0, 3)
	if len(is.tableName) > 0 {
		colIdx := is.valueIndex
		if len(is.columnNames) > 0 {
			colIdx = colIdx % len(is.columnNames)
		}
		if len(is.columnNames) > 0 {
			colName := is.columnNames[colIdx]
			names = append(names, colName)
			names = append(names, fmt.Sprintf("%s.%s", is.tableName, colName))
		}
		names = append(names, fmt.Sprintf("%s.%d", is.tableName, colIdx))
	}

	return names
}
