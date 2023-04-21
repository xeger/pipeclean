package mysql

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
