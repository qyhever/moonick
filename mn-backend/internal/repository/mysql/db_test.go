package mysql

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenDBRequiresConfiguredDSN(t *testing.T) {
	t.Setenv("MOONICK_DATABASE_MYSQL_ADDR", "")
	t.Setenv("MOONICK_DATABASE_MYSQL_USER", "")
	t.Setenv("MOONICK_DATABASE_MYSQL_PASSWORD", "")
	t.Setenv("MOONICK_DATABASE_MYSQL_DB_NAME", "")

	db, err := OpenDB("")
	if err == nil {
		t.Fatal("expected error when DSN is empty")
	}
	if db != nil {
		t.Fatal("expected nil db when DSN is empty")
	}
}

func TestMySQLDriverRegistered(t *testing.T) {
	for _, driver := range sql.Drivers() {
		if driver == "mysql" {
			return
		}
	}
	t.Fatal("expected mysql driver to be registered")
}

func TestSeedAdminUsesDisplayNameColumn(t *testing.T) {
	if !strings.Contains(seedAdminUpsertQuery, "display_name") {
		t.Fatalf("expected seed query to use display_name column, got %q", seedAdminUpsertQuery)
	}
	if strings.Contains(seedAdminUpsertQuery, " name,") {
		t.Fatalf("seed query should not use name column, got %q", seedAdminUpsertQuery)
	}
}

func TestInitSQLMatchesP1Schema(t *testing.T) {
	sqlPath := filepath.Join("..", "..", "..", "docs", "sql", "001_init.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read init sql: %v", err)
	}

	sqlText := string(content)
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS trip_favorites",
		"display_name",
		"publisher_user_id",
		"departure_date",
		"departure_time",
		"price_amount",
		"remark",
		"closed_reason",
		"deleted_at",
		"remark TEXT NOT NULL DEFAULT ''",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("expected init sql to contain %q", fragment)
		}
	}

	if strings.Contains(sqlText, "CREATE TABLE IF NOT EXISTS favorites") {
		t.Fatal("init sql should not create favorites table")
	}
}
