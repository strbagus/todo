package database_test

import (
	"testing"
	"todo/internal/database"
)

func TestInitDB_And_Migrate(t *testing.T) {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='items'").Scan(&tableName)
	if err != nil || tableName != "items" {
		t.Fatalf("expected items table to exist, got error: %v, tableName: %s", err, tableName)
	}
}
