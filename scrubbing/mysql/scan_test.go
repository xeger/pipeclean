package mysql_test

import (
	"reflect"
	"testing"

	"github.com/xeger/pipeclean/scrubbing/mysql"
)

func scan(input string) *mysql.ScrubContext {
	ctx := mysql.NewScrubContext()
	ctx.Scan(input)
	return ctx
}

func TestScanCreateTables(t *testing.T) {
	input := read(t, "create_tables.sql")
	ctx := scan(input)

	expected := map[string][]string{
		"ar_internal_metadata": {"key", "value", "created_at", "updated_at"},
	}
	if !reflect.DeepEqual(ctx.TableColumns, expected) {
		t.Errorf("TableColumns scan failed: expected %v, got %v", expected, ctx.TableColumns)
	}
}
