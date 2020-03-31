package db

import (
	"net"

	"github.com/jinzhu/gorm"
)

// IPAddress is the list of addresses that can be allocated
type IPAddress struct {
	gorm.Model

	Address   string
	Allocated bool
}

// IP returns the parsed, typed IP made for a ipv4 network.
func (ip *IPAddress) IP() net.IP {
	return net.ParseIP(ip.Address).To4()
}

// MACAddress is a network hardware address.
type MACAddress struct {
	gorm.Model

	HardwareAddress string
}

// HardwareAddr returns the typed hardware address for the mac.
func (mac *MACAddress) HardwareAddr() (net.HardwareAddr, error) {
	return net.ParseMAC(mac.HardwareAddress)
}

// Lease is a pre-programmed DHCP lease
type Lease struct {
	gorm.Model

	MAC     MACAddress
	IP      IPAddress
	Dynamic bool
}
