package proto

import (
	"context"
	fmt "fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/erikh/ldhcpd/db"
	"github.com/erikh/ldhcpd/testutil"
	empty "github.com/golang/protobuf/ptypes/empty"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
)

var (
	invalidIPs = []string{
		"1",
		"a.b.c.d",
		"512.512.512.512",
	}

	invalidMacs = []string{
		"1",
		"ab:cd:ef:gh:ji:kl",
		":::::",
	}

	validIPs = []string{
		"1.2.3.4",
		"255.255.255.255",
		"0.0.0.0",
	}

	validMacs = []string{
		"00:00:00:00:00:00",
		"ab:cd:ef:01:02:03",
		"ff:ff:ff:ff:ff:ff",
		"FF:FF:FF:FF:FF:ee",
	}
)

func setupTest(t *testing.T) (LeaseControlClient, net.Listener, *grpc.Server, *db.DB) {
	db, err := db.NewDB("test.db")
	if err != nil {
		t.Fatalf("Error initializing db: %v", err)
	}

	s := Boot(db)
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Error initializing db: %v", err)
	}
	go s.Serve(l)

	cc, err := grpc.Dial(fmt.Sprintf("localhost:%d", l.Addr().(*net.TCPAddr).Port), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Error dialing service: %v", err)
	}

	return NewLeaseControlClient(cc), l, s, db
}

func cleanupTest(t *testing.T, l net.Listener, s *grpc.Server, db *db.DB) {
	s.GracefulStop()
	l.Close()
	db.Close()
	os.Remove("test.db")
}

func TestLeaseHandlerInput(t *testing.T) {
	client, l, s, db := setupTest(t)
	defer cleanupTest(t, l, s, db)

	list, err := client.ListLeases(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Error reading empty list of leases: %v", err)
	}

	if len(list.List) != 0 {
		t.Fatal("List contains data -- it shouldn't")
	}

	for _, badAddress := range invalidMacs {
		_, err := client.GetLease(context.Background(), &MACAddress{Address: badAddress})
		if err == nil {
			t.Fatalf("Did not error on the following address: %v", badAddress)
		}
	}

	// test bad macs and valid ips for set case; then test the inverse.
	for _, badAddress := range invalidMacs {
		for _, ip := range validIPs {
			_, err := client.SetLease(context.Background(), &Lease{MACAddress: badAddress, IPAddress: ip, LeaseEnd: &timestamp.Timestamp{Seconds: time.Now().Add(time.Minute).Unix()}})
			if err == nil {
				t.Fatalf("Did not error with invalid input: bad mac: %v, good ip: %v", badAddress, ip)
			}
		}
	}

	// test bad macs and valid ips for set case; then test the inverse.
	for _, mac := range validMacs {
		for _, badIP := range invalidIPs {
			_, err := client.SetLease(context.Background(), &Lease{MACAddress: mac, IPAddress: badIP, LeaseEnd: &timestamp.Timestamp{Seconds: time.Now().Add(time.Minute).Unix()}})
			if err == nil {
				t.Fatalf("Did not error with invalid input: good mac: %v, bad ip: %v", mac, badIP)
			}
		}
	}

	list, err = client.ListLeases(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Error reading empty list of leases: %v", err)
	}

	if len(list.List) != 0 {
		t.Fatal("List contains data -- it shouldn't")
	}
}

func TestLeaseHandlerSetGetLease(t *testing.T) {
	client, l, s, db := setupTest(t)
	defer cleanupTest(t, l, s, db)

	list, err := client.ListLeases(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Error reading empty list of leases: %v", err)
	}

	if len(list.List) != 0 {
		t.Fatal("List contains data -- it shouldn't")
	}

	table := map[string]string{}
	ipUniq := map[string]struct{}{}
	leaseEnd := time.Now().Add(time.Minute).Unix()

	for i := 0; i < 1000; i++ {
	retry:
		mac := testutil.RandomMAC().String()
		ip := testutil.RandomIP().String()
		if _, ok := table[mac]; ok {
			fmt.Println("mac table collision detected; resetting parameters")
			goto retry
		}
		if _, ok := ipUniq[ip]; ok {
			fmt.Println("ip table collision detected; resetting parameters")
			goto retry
		}
		table[mac] = ip
		ipUniq[ip] = struct{}{}

		_, err := client.SetLease(
			context.Background(),
			&Lease{
				MACAddress: mac,
				IPAddress:  ip,
				LeaseEnd: &timestamp.Timestamp{
					Seconds: leaseEnd,
				},
				LeaseGraceEnd: &timestamp.Timestamp{
					Seconds: leaseEnd,
				},
			})
		if err != nil {
			t.Fatalf("Error while inserting mac/ip %s/%s: %v", mac, ip, err)
		}
	}

	for mac, ip := range table {
		lease, err := client.GetLease(context.Background(), &MACAddress{Address: mac})
		if err != nil {
			t.Fatalf("Error locating entered mac for address %s: %v", mac, err)
		}

		if lease.IPAddress != ip {
			t.Fatalf("IPs do not match for mac lease: %s/%s", lease.IPAddress, ip)
		}

		if lease.LeaseEnd.Seconds != leaseEnd {
			t.Fatalf("Lease ending times do not match: %v/%v", lease.LeaseEnd.Seconds, leaseEnd)
		}
	}

	list, err = client.ListLeases(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("error listing leases: %v", err)
	}

	if len(list.List) != 1000 {
		t.Fatalf("List did not return all data: only %d rows exist or were returned", len(list.List))
	}

	for _, lease := range list.List {
		if len(table[lease.MACAddress]) == 0 || table[lease.MACAddress] != lease.IPAddress {
			t.Errorf("lease does not exist in list, or there is a mismatch")
		}
	}

	for _, lease := range list.List {
		if _, err := client.RemoveLease(context.Background(), &MACAddress{Address: lease.MACAddress}); err != nil {
			t.Fatalf("error removing existing lease for %s: %v", lease.MACAddress, err)
		}
	}

	for _, lease := range list.List {
		if _, err := client.RemoveLease(context.Background(), &MACAddress{Address: lease.MACAddress}); err == nil {
			t.Fatalf("no error removing missing deleted for %s: %v", lease.MACAddress, err)
		}
	}
}
