package dhcpd

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
	"github.com/insomniacslk/dhcp/rfc1035label"
	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
)

func TestRealClientACK(t *testing.T) {
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

	term := setupDHCPHandler(t, config)
	defer os.Remove(config.DBFile)
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

func TestParallelAcquisition(t *testing.T) {
	bridge := setupTest(t)
	defer cleanupTest(t)

	time.Sleep(time.Second)

	config := Config{
		Lease: Lease{
			Duration: 5 * time.Second,
		},
		SearchDomains: []string{"internal"},
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

	term := setupDHCPHandler(t, config)
	defer os.Remove(config.DBFile)
	defer close(term)

	from, to := config.DynamicRange.Dimensions()
	errChan := make(chan error, 50)
	doneChan := make(chan struct{}, 50)

	for i := 0; i < 50; i++ {
		go func(i int) {
			defer func() { doneChan <- struct{}{} }()

			linkName := fmt.Sprintf("pclient%d", i)
			addVethPair(linkName, bridge)
			defer teardownLink(linkName)

			c, err := nclient4.New(linkName)
			if err != nil {
				errChan <- err
				return
			}

			offer, ack, err := c.Request(context.Background())
			if err != nil {
				errChan <- errors.Wrap(err, "could not complete request")
				return
			}

			if !offer.YourIPAddr.Equal(ack.YourIPAddr) {
				errChan <- errors.Wrap(err, "IPs between offer and ack are not equal")
				return
			}

			if !dhcp4.IPInRange(from, to, offer.YourIPAddr) {
				errChan <- errors.Wrap(err, "issued IP not in range")
				return
			}

			if !dhcp4.IPInRange(from, to, ack.YourIPAddr) {
				errChan <- errors.Wrap(err, "issued IP not in range")
				return
			}

			labels, err := rfc1035label.FromBytes(offer.GetOneOption(dhcpv4.OptionDNSDomainSearchList))
			if err != nil {
				errChan <- err
				return
			}

			if len(labels.Labels) == 0 {
				errChan <- errors.Wrap(err, "search list was not provided in offer")
				return
			}

			if labels.Labels[0] != "internal" {
				errChan <- errors.Wrap(err, "specified domain name is not present")
				return
			}
		}(i)
	}

	for i := 0; i < 50; i++ {
		<-doneChan
	}

	select {
	case err := <-errChan:
		dumpInterfaces()
		t.Fatal(err)
	default:
	}
}
