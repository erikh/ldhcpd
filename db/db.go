package db

import (
	"net"

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

	return &DB{db: db}, db.AutoMigrate(&IPAddress{}, &MACAddress{}, &Lease{}).Error
}

// Close the database
func (db *DB) Close() error {
	return db.db.Close()
}

// AddAddress adds an address to the database if it does not already exist.
func (db *DB) AddAddress(ip net.IP) error {
	tx := db.db.Begin()
	defer tx.Rollback()

	return tx.FirstOrCreate(&IPAddress{Address: ip.String()}, "address = ?", ip.String()).Commit().Error
}
