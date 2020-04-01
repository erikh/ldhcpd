package proto

import (
	"context"
	fmt "fmt"
	"net"
	"os"
	"testing"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/db"
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
