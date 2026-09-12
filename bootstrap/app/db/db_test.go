package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type transactionTestRow struct {
	ID uint `gorm:"primaryKey"`
	Name string
}

func TestMigrateAndTransactionHelpers(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite3")
	t.Setenv("DB_NAME", ":memory:")

	err := Initialize()
	assert.NoError(t, err)
	defer func() { _ = Close() }()

	err = Migrate(&transactionTestRow{})
	assert.NoError(t, err)

	err = InTransaction(func(tx *gorm.DB) error {
		return tx.Create(&transactionTestRow{Name: "alpha"}).Error
	})
	assert.NoError(t, err)

	var count int64
	err = Get().Model(&transactionTestRow{}).Count(&count).Error
	assert.NoError(t, err)
	assert.EqualValues(t, 1, count)

	err = InTransaction(func(tx *gorm.DB) error {
		return tx.Create(&transactionTestRow{Name: "beta"}).Error
	})
	assert.NoError(t, err)

	err = InTransaction(func(tx *gorm.DB) error {
		return tx.Create(&transactionTestRow{Name: "gamma"}).Error
	})
	assert.NoError(t, err)

	var total int64
	err = Get().Model(&transactionTestRow{}).Count(&total).Error
	assert.NoError(t, err)
	assert.EqualValues(t, 3, total)
}

func TestInTransactionRollsBackOnError(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite3")
	t.Setenv("DB_NAME", ":memory:")

	err := Initialize()
	assert.NoError(t, err)
	defer func() { _ = Close() }()

	err = Migrate(&transactionTestRow{})
	assert.NoError(t, err)

	err = InTransaction(func(tx *gorm.DB) error {
		if err := tx.Create(&transactionTestRow{Name: "keep"}).Error; err != nil {
			return err
		}
		return assert.AnError
	})
	assert.Error(t, err)

	var count int64
	err = Get().Model(&transactionTestRow{}).Count(&count).Error
	assert.NoError(t, err)
	assert.EqualValues(t, 0, count)
}
