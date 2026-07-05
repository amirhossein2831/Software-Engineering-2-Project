// Package database provides the shared GORM/Postgres connection helper. Each
// service opens its own database (database-per-service) with a DSN from env.
package database

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects to Postgres using the given DSN and returns a *gorm.DB.
func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, errors.New("database: empty DSN")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		TranslateError:         true, // map driver errors to gorm.ErrDuplicatedKey etc.
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// MustOpen is like Open but panics on error. Use at service startup.
func MustOpen(dsn string) *gorm.DB {
	db, err := Open(dsn)
	if err != nil {
		panic(err)
	}
	return db
}
