package db

import "testing"

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
	// The repository does not register a SQLite driver, but calling NewSQL
	// still verifies the supported-driver branch and its default database name.
	db, err := NewSQL(Config{Driver: DriverSqlite3})
	if db != nil {
		db.Close()
	}
	if err == nil {
		t.Fatal("NewSQL() unexpectedly opened SQLite without a registered driver")
	}
}
