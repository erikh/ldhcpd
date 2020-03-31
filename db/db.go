package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // import sqlite3
	"github.com/pkg/errors"
)

// DB is the outer shell for the gorm DB handle.
type DB struct {
	db *gorm.DB
}

// NewDB opens the DB
func NewDB(dbfile string) (*DB, error) {
	db, err := gorm.Open("sqlite3", dbfile)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to db")
	}

	if err := db.AutoMigrate(&Lease{}).Error; err != nil {
		return nil, errors.Wrap(err, "while migrating database")
	}

	return &DB{db: db}, nil
}

// Close the database
func (db *DB) Close() error {
	return db.db.Close()
}
