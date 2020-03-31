package db

import (
	"net"
	"os"
	"testing"
)

func TestDBIPAddress(t *testing.T) {
	db, err := NewDB("test.db")
	if err != nil {
		t.Fatalf("Could not open test database: %v", err)
	}
	defer db.Close()
	defer os.Remove("test.db")

	if _, err := db.FindAddress(net.ParseIP("1.2.3.4")); err == nil {
		t.Fatalf("found address 1.2.3.4 when should not be found")
	}

	for i := 0; i < 10; i++ {
		if err := db.AddAddress(net.ParseIP("1.2.3.4")); err != nil {
			t.Fatalf("Could not add address: %v", err)
		}
	}

	address, err := db.FindAddress(net.ParseIP("1.2.3.4"))
	if err != nil {
		t.Fatalf("did not find address 1.2.3.4 when should be found")
	}

	if address.IP().String() != "1.2.3.4" {
		t.Fatalf("IP address did not match")
	}
}
