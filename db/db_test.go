package db

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewSQLRejectsUnsupportedDriver(t *testing.T) {
	db, err := NewSQL(Config{Driver: DriverMysql})
	if db != nil {
		t.Fatal("NewSQL() returned a database for an unsupported driver")
	}
	if err == nil {
		t.Fatal("NewSQL() returned nil error for an unsupported driver")
	}
}

func TestNewSQLUsesSQLiteBranch(t *testing.T) {
	db, err := NewSQL(Config{Driver: DriverSqlite3})
	if err != nil {
		t.Fatalf("NewSQL() returned an unexpected error for sqlite3: %v", err)
	}
	if db == nil {
		t.Fatal("NewSQL() returned nil database for sqlite3")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("database close failed: %v", err)
	}
}

func TestValidateConfigRejectsUnsupportedDriver(t *testing.T) {
	if err := ValidateConfig(Config{Driver: DriverMysql}); err == nil {
		t.Fatal("ValidateConfig() accepted an unsupported driver")
	}
}

func TestWithTransactionCommitsAndRollsBack(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() failed: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	if err := WithTransaction(db, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO items (name) VALUES ('alpha')")
		return err
	}); err != nil {
		t.Fatalf("WithTransaction() commit path failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		t.Fatalf("QueryRow() failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after commit, got %d", count)
	}

	if err := WithTransaction(db, func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO items (name) VALUES ('beta')"); err != nil {
			return err
		}
		return errors.New("rollback me")
	}); err == nil {
		t.Fatal("WithTransaction() unexpectedly committed a failing transaction")
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		t.Fatalf("Verify rollback query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected rollback to keep 1 row, got %d", count)
	}
}

func TestRepositoryRoundTrip(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() failed: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE person_rows (id INTEGER PRIMARY KEY, name TEXT NOT NULL, age INTEGER)"); err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	type personRow struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		Age  int    `db:"age"`
	}

	repo := NewRepository[personRow](db, "person_rows")

	createdID, err := repo.Create(personRow{Name: "Ada", Age: 36})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if createdID == 0 {
		t.Fatal("Create() returned an invalid ID")
	}

	items, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item after Create(), got %d", len(items))
	}
	if items[0].Name != "Ada" || items[0].Age != 36 {
		t.Fatalf("unexpected created row: %+v", items[0])
	}

	stored, err := repo.FindByID(createdID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}
	if stored.Name != "Ada" {
		t.Fatalf("FindByID() returned unexpected row: %+v", stored)
	}

	if err := repo.UpdateByID(createdID, map[string]any{"name": "Grace", "age": 37}); err != nil {
		t.Fatalf("UpdateByID() failed: %v", err)
	}

	updated, err := repo.FindByID(createdID)
	if err != nil {
		t.Fatalf("FindByID() after update failed: %v", err)
	}
	if updated.Name != "Grace" || updated.Age != 37 {
		t.Fatalf("unexpected updated row: %+v", updated)
	}

	if err := repo.DeleteByID(createdID); err != nil {
		t.Fatalf("DeleteByID() failed: %v", err)
	}

	remaining, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() after delete failed: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no rows after delete, got %d", len(remaining))
	}
}
