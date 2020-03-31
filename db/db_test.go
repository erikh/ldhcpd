package db

import (
	"net"
	"os"
	"testing"
	"time"
)

const (
	fakeMAC  = "01:01:01:01:01:01"
	fakeMAC2 = "02:02:02:02:02:02"
)

func TestDBLease(t *testing.T) {
	db, err := NewDB("test.db")
	if err != nil {
		t.Fatalf("Could not open test database: %v", err)
	}
	defer db.Close()
	defer os.Remove("test.db")

	mac, err := net.ParseMAC(fakeMAC)
	if err != nil {
		t.Fatalf("Could not parse fake mac %v", fakeMAC)
	}

	mac2, err := net.ParseMAC(fakeMAC2)
	if err != nil {
		t.Fatalf("Could not parse fake mac %v", fakeMAC2)
	}

	if _, err := db.GetLease(mac); err == nil {
		t.Fatalf("Found lease where there shouldn't be one")
	}

	if err := db.SetLease(mac, net.ParseIP("10.0.0.1"), false, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Found lease where there shouldn't be one")
	}

	l, err := db.GetLease(mac)
	if err != nil {
		t.Fatalf("Did not find lease where there should be one")
	}

	if l.IP().String() != "10.0.0.1" {
		t.Fatalf("IP (%v) was not equal to 10.0.0.1", l.IPAddress)
	}

	if err := db.SetLease(mac2, net.ParseIP("10.0.0.1"), false, time.Now().Add(time.Hour)); err == nil {
		t.Fatal("Should not have been able to create a second lease for 10.0.0.1")
	}

	if err := db.SetLease(mac, net.ParseIP("10.0.0.2"), false, time.Now().Add(time.Hour)); err == nil {
		t.Fatalf("Should not have been able to create a second lease for %v", mac.String())
	}
}
