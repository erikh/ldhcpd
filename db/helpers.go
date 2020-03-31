package db

import (
	"net"

	"github.com/jinzhu/gorm"
)

// AddAddress adds an address to the database if it does not already exist.
func (db *DB) AddAddress(ip net.IP) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		var count int
		err := tx.Model(&IPAddress{}).Where("address = ?", ip.String()).Count(&count).Error
		if count == 0 || gorm.IsRecordNotFoundError(err) {
			return tx.Create(&IPAddress{Address: ip.String()}).Error
		} else if err != nil {
			return err
		}

		return nil
	})
}

// FindAddress finds the address if it exists.
func (db *DB) FindAddress(ip net.IP) (*IPAddress, error) {
	address := &IPAddress{}
	return address, db.db.Transaction(func(tx *gorm.DB) error {
		return tx.First(address, "address = ?", ip.String()).Error
	})
}
