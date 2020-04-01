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

func getBytes(size int) []byte {
	byt := make([]byte, size)
	n, err := rand.Read(byt)
	if err != nil {
		panic(err)
	}

	if n != size {
		panic("short read in randommac")
	}

	return byt
}

// RandomMAC uses entropy to generate a mac... or else
func RandomMAC() net.HardwareAddr {
	return net.HardwareAddr(getBytes(6))
}

// RandomIP returns a random IP.
func RandomIP() net.IP {
	return net.IP(getBytes(4))
}
