package testutil

import (
	"crypto/rand"
	"net"
)

var (
	// FakeMAC is testing data
	FakeMAC = RandomMAC()
	// FakeMAC2 is testing data
	FakeMAC2 = RandomMAC()
)

// RandomMAC uses entropy to generate a mac... or else
func RandomMAC() net.HardwareAddr {
	hwaddr := make([]byte, 6)
	n, err := rand.Read(hwaddr)
	if err != nil {
		panic(err)
	}

	if n != 6 {
		panic("short read in randommac")
	}

	return net.HardwareAddr(hwaddr)
}
