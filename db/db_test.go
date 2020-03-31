package db

import (
	"bytes"
	"net"
	"os"
	"testing"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/testutil"
)

func TestDBLeaseCRUD(t *testing.T) {
	db, err := NewDB("test.db")
	if err != nil {
		t.Fatalf("Could not open test database: %v", err)
	}
	defer db.Close()
	defer os.Remove("test.db")

	if err := db.SetLease(testutil.FakeMAC, net.ParseIP("10.0.0.1"), false, time.Now().Add(time.Second)); err != nil {
		t.Fatalf("could not set basic lease: %v", err)
	}

	if err := db.SetLease(testutil.FakeMAC2, net.ParseIP("10.0.0.2"), false, time.Now().Add(time.Second)); err != nil {
		t.Fatalf("could not set basic lease: %v", err)
	}

	time.Sleep(time.Second)

	count, err := db.PurgeLeases()
	if err != nil {
		t.Fatalf("could not purge leases: %v", err)
	}

	if count != 2 {
		t.Fatalf("Did not purge the right number of leases, expected 2, got %d", count)
	}

	if err := db.SetLease(testutil.FakeMAC, net.ParseIP("10.0.0.1"), false, time.Now().Add(time.Second)); err != nil {
		t.Fatalf("could not set basic lease: %v", err)
	}

	if _, err := db.RenewLease(testutil.FakeMAC2, time.Now().Add(time.Minute)); err == nil {
		t.Fatal("did not error renewing lease for missing mac")
	}

	lease, err := db.RenewLease(testutil.FakeMAC, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("could not renew lease: %v", err)
	}

	// if the time was only still a second, subtracting it would yield a time
	// before the present since at least a nanosecond will have passed during the
	// test.
	if lease.LeaseEnd.Add(-time.Second).Before(time.Now()) {
		t.Fatal("Lease ending was not updated")
	}

	if err := db.SetLease(testutil.FakeMAC2, net.ParseIP("10.0.0.2"), false, time.Now().Add(time.Second)); err != nil {
		t.Fatalf("could not set basic lease: %v", err)
	}
}

func TestDBLease(t *testing.T) {
	db, err := NewDB("test.db")
	if err != nil {
		t.Fatalf("Could not open test database: %v", err)
	}
	defer db.Close()
	defer os.Remove("test.db")

	if _, err := db.GetLease(testutil.FakeMAC); err == nil {
		t.Fatalf("Found lease where there shouldn't be one")
	}

	if err := db.SetLease(testutil.FakeMAC, net.ParseIP("10.0.0.1"), false, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Found lease where there shouldn't be one")
	}

	l, err := db.GetLease(testutil.FakeMAC)
	if err != nil {
		t.Fatalf("Did not find lease where there should be one")
	}

	if l.IP().String() != "10.0.0.1" {
		t.Fatalf("IP (%v) was not equal to 10.0.0.1", l.IPAddress)
	}

	tmpMac, err := l.HardwareAddr()
	if err != nil {
		t.Fatalf("While parsing mac for lease: %v", err)
	}

	if !bytes.Equal(tmpMac, testutil.FakeMAC) {
		t.Fatalf("Mac address is not equal in lease: %v", tmpMac.String())
	}

	if err := db.SetLease(testutil.FakeMAC2, net.ParseIP("10.0.0.1"), false, time.Now().Add(time.Hour)); err == nil {
		t.Fatal("Should not have been able to create a second lease for 10.0.0.1")
	}

	if err := db.SetLease(testutil.FakeMAC, net.ParseIP("10.0.0.2"), false, time.Now().Add(time.Hour)); err == nil {
		t.Fatalf("Should not have been able to create a second lease for %v", testutil.FakeMAC.String())
	}
}
