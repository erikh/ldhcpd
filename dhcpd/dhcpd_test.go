package dhcpd

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestBasicACK(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	time.Sleep(time.Second)

	config := Config{
		Lease: Lease{
			Duration: 5 * time.Second,
		},
		DNSServers: []string{
			"10.0.0.1",
			"1.1.1.1",
		},
		Gateway: "10.0.20.1",
		DynamicRange: Range{
			From: "10.0.20.50",
			To:   "10.0.20.100",
		},
		DBFile: "test.db",
	}

	handler, term := setupDHCPHandler(t, config)
	defer os.Remove(config.DBFile)
	defer handler.Close()
	defer close(term)

	ip := testDHCP(t)
	if !ip.Equal(net.ParseIP("10.0.20.50")) {
		dumpInterfaces()
		t.Fatalf("Was not the expected ip: was %v", ip)
	}

	ip = testDHCP(t)
	if !ip.Equal(net.ParseIP("10.0.20.50")) {
		dumpInterfaces()
		t.Fatalf("Was not the expected ip: was %v", ip)
	}

	time.Sleep(6 * time.Second) // ensure lease is expired

	// I know this is especially shitty behavior and could be done better. I will
	// add a shadow lease table eventually, just making this note so if anyone
	// reads this code they don't think I'm a massive dimwit (at least for this
	// reason). This is just a simpler solution; maintain the lease, or lose it.
	ip = testDHCP(t)
	if !ip.Equal(net.ParseIP("10.0.20.51")) {
		dumpInterfaces()
		t.Fatalf("Was not the expected IP: was %v", ip)
	}
}
