package postgres

import (
	"reflect"
	"testing"
)

func TestBuilderUsesPostgresPlaceholders(t *testing.T) {
	t.Parallel()

	query, args, err := Builder.
		Select("id", "name").
		From("manufactures").
		Where("id = ? AND deleted_at IS NULL", "3422b448-2460-4fd2-9183-8000de6f8343").
		ToSql()
	if err != nil {
		t.Fatalf("ToSql() error = %v", err)
	}

	const wantQuery = "SELECT id, name FROM manufactures WHERE id = $1 AND deleted_at IS NULL"
	if query != wantQuery {
		t.Fatalf("query = %q, want %q", query, wantQuery)
	}
	if wantArgs := []any{"3422b448-2460-4fd2-9183-8000de6f8343"}; !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", args, wantArgs)
	}
}
