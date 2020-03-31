package testutil

import "net"

const (
	fakeMACAddr  = "01:01:01:01:01:01"
	fakeMACAddr2 = "02:02:02:02:02:02"
)

var FakeMAC, FakeMAC2 net.HardwareAddr

func init() {
	var err error

	FakeMAC, err = net.ParseMAC(fakeMACAddr)
	if err != nil {
		panic(err)
	}

	FakeMAC2, err = net.ParseMAC(fakeMACAddr2)
	if err != nil {
		panic(err)
	}
}
