package db

import (
	"net"
	"time"

	"github.com/jinzhu/gorm"
)

// Lease is a pre-programmed DHCP lease
type Lease struct {
	MACAddress    string `gorm:"primary_key"`
	IPAddress     string `gorm:"unique"`
	Dynamic       bool
	LeaseEnd      time.Time
	LeaseGraceEnd time.Time
	Persistent    bool
}

// IP returns the parsed, typed IP made for a ipv4 network.
func (l *Lease) IP() net.IP {
	return net.ParseIP(l.IPAddress).To4()
}

// HardwareAddr returns the typed hardware address for the mac.
func (l *Lease) HardwareAddr() (net.HardwareAddr, error) {
	return net.ParseMAC(l.MACAddress)
}

// GetLease retrieves the lease if possible, otherwise returns error.
func (db *DB) GetLease(mac net.HardwareAddr) (*Lease, error) {
	l := &Lease{}

	return l, db.db.Transaction(func(tx *gorm.DB) error {
		return tx.First(l, "mac_address = ?", mac.String()).Error
	})
}

// SetLease creates a lease if possible.
func (db *DB) SetLease(mac net.HardwareAddr, ip net.IP, dynamic, persistent bool, end, graceEnd time.Time) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&Lease{
			MACAddress:    mac.String(),
			IPAddress:     ip.String(),
			Dynamic:       dynamic,
			LeaseEnd:      end,
			LeaseGraceEnd: graceEnd,
			Persistent:    persistent,
		}).Error
	})
}

// RenewLease renews a lease up to the given time.
func (db *DB) RenewLease(mac net.HardwareAddr, end, graceEnd time.Time) (*Lease, error) {
	l := &Lease{}

	return l, db.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(l, "mac_address = ?", mac.String()).Error; err != nil {
			return err
		}

		l.LeaseEnd = end
		l.LeaseGraceEnd = graceEnd
		return tx.Save(l).Error
	})
}

// RemoveLease removes a lease based on MAC.
func (db *DB) RemoveLease(mac net.HardwareAddr) error {
	count := int64(0)
	err := db.db.Transaction(func(tx *gorm.DB) error {
		// shadowing db
		db := tx.Delete(&Lease{}, "mac_address = ?", mac.String())
		count = db.RowsAffected
		return db.Error
	})

	if err == nil && count == 0 {
		return gorm.ErrRecordNotFound
	}

	return err
}

// PurgeLeases removes all leases that are expired. It returns the count of expired leases, and an error if any.
func (db *DB) PurgeLeases(ignoreGrace bool) (int64, error) {
	var rows int64
	return rows, db.db.Transaction(func(tx *gorm.DB) error {
		// shadowing db
		now := time.Now()
		var db *gorm.DB
		if ignoreGrace { // we need ips
			db = tx.Delete(&Lease{}, "lease_end < ? and not persistent", now)
		} else {
			db = tx.Delete(&Lease{}, "lease_end < ? and lease_grace_end < ? and not persistent", now, now)
		}

		rows = db.RowsAffected
		return db.Error
	})
}

// ListLeases returns all leases in the lease table.
func (db *DB) ListLeases() ([]*Lease, error) {
	leases := []*Lease{}

	return leases, db.db.Transaction(func(tx *gorm.DB) error {
		return tx.Find(&leases).Error
	})
}
