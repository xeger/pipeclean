package scrubbing

import (
	"bytes"
	"strings"

	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

// Turns an AST back into a string.
func restore(stmt ast.StmtNode) string {
	buf := new(bytes.Buffer)
	ctx := format.NewRestoreCtx(format.RestoreKeyWordUppercase|format.RestoreNameBackQuotes|format.RestoreStringSingleQuotes|format.RestoreStringWithoutDefaultCharset, buf)
	err := stmt.Restore(ctx)
	if err != nil {
		panic(err)
	}
	s := buf.String()
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s + ";\n"
}
