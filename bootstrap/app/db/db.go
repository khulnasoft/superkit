package db

import (
	"os"
	"time"

	"github.com/khulnasoft/superkit/db"

	_ "github.com/mattn/go-sqlite3"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	dbInstance *gorm.DB
	initialized bool
)

// IsReady reports whether the database is initialized and responding to pings.
func IsReady() bool {
	if !initialized || dbInstance == nil {
		return false
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return false
	}

	return sqlDB.Ping() == nil
}

// Get returns the instantiated DB instance. Panics if called before Initialize.
func Get() *gorm.DB {
	if !initialized {
		panic("db: Get() called before Initialize()")
	}
	return dbInstance
}

// Initialize sets up the database connection from environment configuration.
func Initialize() error {
	config := db.Config{
		Driver:   os.Getenv("DB_DRIVER"),
		Name:     os.Getenv("DB_NAME"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		Host:     os.Getenv("DB_HOST"),
	}

	dbinst, err := db.NewSQL(config)
	if err != nil {
		initialized = false
		dbInstance = nil
		return err
	}

	dbinst.SetMaxOpenConns(5)
	dbinst.SetMaxIdleConns(2)
	dbinst.SetConnMaxLifetime(30 * time.Minute)

	if _, err := dbinst.Exec("PRAGMA journal_mode=WAL"); err != nil {
		initialized = false
		dbInstance = nil
		return err
	}
	if _, err := dbinst.Exec("PRAGMA busy_timeout=5000"); err != nil {
		initialized = false
		dbInstance = nil
		return err
	}

	switch config.Driver {
	case db.DriverSqlite3:
		dbInstance, err = gorm.Open(sqlite.New(sqlite.Config{
			Conn: dbinst,
		}))
	default:
		initialized = false
		dbInstance = nil
		return err
	}

	if err != nil {
		initialized = false
		dbInstance = nil
		return err
	}

	initialized = true
	return nil
}

// Close closes the database connection.
func Close() error {
	if !initialized || dbInstance == nil {
		return nil
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return err
	}

	initialized = false
	dbInstance = nil
	return sqlDB.Close()
}

func Migrate(models ...interface{}) error {
	if !initialized || dbInstance == nil {
		return nil
	}
	return dbInstance.AutoMigrate(models...)
}

func InTransaction(fn func(tx *gorm.DB) error) error {
	if !initialized || dbInstance == nil {
		return nil
	}
	return dbInstance.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
