package dhcpd

import (
	"net"
	"os"
	"testing"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/testutil"
)

func TestAllocator(t *testing.T) {
	config := Config{
		Lease: Lease{
			Duration: 100 * time.Millisecond,
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

	time.Sleep(100 * time.Millisecond) // lease duration

	count, err := db.PurgeLeases()
	if err != nil {
		t.Fatalf("could not purge leases: %v", err)
	}

	if count != 2 {
		t.Fatal("Did not purge all leases!")
	}

	if _, err := a.Allocate(testutil.FakeMAC, false); err != nil {
		t.Fatalf("error allocating first ip: %v", err)
	}

	if _, err := a.Allocate(testutil.FakeMAC2, false); err != nil {
		t.Fatalf("Could not allocate second mac: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := a.Allocate(testutil.FakeMAC, true); err != nil {
		t.Fatalf("error allocating first ip: %v", err)
	}

	if _, err := a.Allocate(testutil.FakeMAC2, true); err != nil {
		t.Fatalf("Could not allocate second mac: %v", err)
	}

	count, err = db.PurgeLeases()
	if err != nil {
		t.Fatalf("could not purge leases: %v", err)
	}

	if count != 0 {
		t.Fatal("purged a renewed lease")
	}
}

func TestAllocatorCycles(t *testing.T) {
	config := Config{
		Lease: Lease{
			Duration: 100 * time.Millisecond,
		},
		DNSServers: []string{
			"10.0.0.1",
			"1.1.1.1",
		},
		Gateway: "10.0.20.1",
		DynamicRange: Range{
			From: "10.0.20.50",
			To:   "10.0.20.50",
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
		t.Fatalf("allocation failed: %v", err)
	}

	if ip.String() != "10.0.20.50" {
		t.Fatal("IP was not allocated properly")
	}

	if _, err := a.Allocate(testutil.FakeMAC2, false); err != ErrRangeExhausted {
		if err != nil {
			t.Logf("Error was: %v", err)
		}

		t.Fatalf("allocation did not fail!")
	}

	time.Sleep(100 * time.Millisecond)

	count, err := db.PurgeLeases()
	if err != nil {
		t.Fatalf("could not purge leases: %v", err)
	}

	if count != 1 {
		t.Fatal("Did not purge all leases!")
	}

	if _, err := a.Allocate(testutil.FakeMAC2, false); err != nil {
		t.Fatalf("Could not allocate against other mac after purge: %v", err)
	}
}

func TestAllocatorGaps(t *testing.T) {
	config := Config{
		Lease: Lease{
			Duration: 100 * time.Millisecond,
		},
		DNSServers: []string{
			"10.0.0.1",
			"1.1.1.1",
		},
		Gateway: "10.0.20.1",
		DynamicRange: Range{
			From: "10.0.20.50",
			To:   "10.0.20.59",
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

	keep := map[string]net.HardwareAddr{}

	for i := 0; i < 10; i++ {
		mac := testutil.RandomMAC()
		ip, err := a.Allocate(mac, false)
		if err != nil {
			t.Fatalf("Allocation failed: %v", err)
		}

		if i%2 == 0 {
			keep[ip.String()] = mac
		}
	}

	time.Sleep(100 * time.Millisecond)

	for ip, mac := range keep {
		newip, err := a.Allocate(mac, true)
		if err != nil {
			t.Fatalf("Error allocating for renewal: %v", err)
		}

		if newip.String() != ip {
			t.Fatalf("Allocation did not reap same ip: %v/%v", newip.String(), ip)
		}
	}

	count, err := db.PurgeLeases()
	if err != nil {
		t.Fatalf("Could not purge leases: %v", err)
	}

	if count != 5 {
		t.Fatalf("Purged n != 5 records: %v", count)
	}

	for i := 0; i < 5; i++ {
		mac := testutil.RandomMAC()
		ip, err := a.Allocate(mac, false)
		if err != nil {
			t.Fatalf("Allocation failed: %v", err)
		}

		if _, ok := keep[ip.String()]; ok {
			t.Fatalf("Re-allocated renewed ip: %v %v %v", mac, ip.String(), keep[ip.String()])
		}

		keep[ip.String()] = mac
	}

	if ip, err := a.Allocate(testutil.RandomMAC(), false); err != ErrRangeExhausted {
		t.Fatalf("range was not exhausted during testing: %v", ip)
	}
}

func TestAllocatorPersistent(t *testing.T) {
	config := Config{
		Lease: Lease{
			Duration: 100 * time.Millisecond,
		},
		DNSServers: []string{
			"10.0.0.1",
			"1.1.1.1",
		},
		Gateway: "10.0.20.1",
		DynamicRange: Range{
			From: "10.0.20.50",
			To:   "10.0.20.59",
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

	mac := testutil.RandomMAC()
	if err := db.SetLease(mac, net.ParseIP("1.2.3.4"), false, true, time.Now()); err != nil {
		t.Fatalf("Error setting lease: %v", err)
	}

	time.Sleep(time.Second)

	count, err := db.PurgeLeases()
	if err != nil {
		t.Fatalf("Error purging leases: %v", err)
	}

	if count != 0 {
		t.Fatal("Purged persistent lease for some reason")
	}

	ip, err := a.Allocate(mac, false)
	if err != nil {
		t.Fatalf("Error allocating mac: %v", err)
	}

	if ip.String() != "1.2.3.4" {
		t.Fatalf("Got wrong ip back from allocator: %v, should be 1.2.3.4", ip.String())
	}
}
