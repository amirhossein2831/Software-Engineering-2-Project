package database

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, errors.New("database: empty DSN")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		TranslateError:         true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MustOpen(dsn string) *gorm.DB {
	db, err := Open(dsn)
	if err != nil {
		panic(err)
	}
	return db
}
