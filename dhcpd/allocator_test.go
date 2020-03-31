package dhcpd

import (
	"os"
	"testing"

	"code.hollensbe.org/erikh/ldhcpd/testutil"
)

func TestAllocator(t *testing.T) {
	config := Config{
		LeaseDuration: defaultLeaseDuration,
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
	defer os.Remove("test.db")

	db, err := config.NewDB()
	if err != nil {
		t.Fatalf("Error creating database: %v", err)
	}
	defer db.Close()

	a, err := NewAllocator(db, config, nil)
	if err != nil {
		t.Fatalf("error creating allocator: %v", err)
	}

	ip, err := a.Allocate(testutil.FakeMAC, false)
	if err != nil {
		t.Fatalf("error allocating first ip: %v", err)
	}

	if ip.String() != config.DynamicRange.From {
		t.Fatalf("Expected allocated ip was incorrect, was %v, supposed to be %v", ip, config.DynamicRange.From)
	}

	ip2, err := a.Allocate(testutil.FakeMAC, false)
	if err != nil {
		t.Fatalf("error allocating first ip: %v", err)
	}

	ip2, err = a.Allocate(testutil.FakeMAC2, false)
	if err != nil {
		t.Fatalf("Could not allocate second mac: %v", err)
	}

	if ip.String() == ip2.String() {
		t.Fatal("Allocator handed out same address twice")
	}
}
