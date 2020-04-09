package dhcpd

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

const defaultBridge = "testbridge0"

var initialInterfaces = map[string]net.IP{"dhcpd0": net.ParseIP("10.0.20.1"), "dhclient0": nil}

func addVethPair(name string, bridge *netlink.Bridge) error {
	peerName := name + "-peer"

	la := netlink.NewLinkAttrs()
	la.Name = name
	veth := &netlink.Veth{LinkAttrs: la, PeerName: peerName}
	if err := netlink.LinkAdd(veth); err != nil {
		return errors.Wrap(err, "could not create veth pair")
	}

	peer, err := netlink.LinkByName(peerName)
	if err != nil {
		return errors.Wrap(err, "could not locate created peer")
	}

	if err := netlink.LinkSetUp(veth); err != nil {
		return errors.Wrap(err, "could not raise peer")
	}

	if err := netlink.LinkSetUp(peer); err != nil {
		return errors.Wrap(err, "could not raise peer")
	}

	if err := netlink.LinkSetMaster(peer, bridge); err != nil {
		return errors.Wrap(err, "could not add peer to bridge")
	}

	return nil
}

func setupTest(t *testing.T) {
	cleanupTest(t)

	la := netlink.NewLinkAttrs()
	la.Name = defaultBridge
	bridge := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(bridge); err != nil {
		t.Fatalf("Could not create bridge: %v", err)
	}

	for name, ip := range initialInterfaces {
		if err := addVethPair(name, bridge); err != nil {
			t.Fatalf("Could not add veth pair: %v", err)
		}

		link, err := netlink.LinkByName(name)
		if err != nil {
			t.Fatalf("Could not find newly added link %v: %v", name, err)
		}

		if ip != nil {
			addr := &netlink.Addr{
				IPNet: &net.IPNet{
					IP:   ip,
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			}

			if err := netlink.AddrAdd(link, addr); err != nil {
				t.Fatalf("Could not configure ip: %v", err)
			}
		}

		if err := netlink.LinkSetUp(link); err != nil {
			t.Fatalf("Could not raise link %v: %v", name, err)
		}
	}

	if err := netlink.LinkSetUp(bridge); err != nil {
		t.Fatalf("Could not enable bridge: %v", err)
	}
}

func cleanupTest(t *testing.T) {
	for name := range initialInterfaces {
		link, err := netlink.LinkByName(name)
		if err == nil {
			netlink.LinkSetDown(link)
			if err := netlink.LinkDel(link); err != nil {
				t.Fatalf("Could not delete existing link %v: %v", name, err)
			}
		}
	}

	bridge, err := netlink.LinkByName(defaultBridge)
	if err == nil {
		if err := netlink.LinkSetDown(bridge); err != nil {
			t.Fatalf("Could not disable bridge: %v", err)
		}

		if err := netlink.LinkDel(bridge); err != nil {
			t.Fatalf("Could not delete bridge: %v", err)
		}
	}
}

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

	db, err := config.NewDB()
	if err != nil {
		t.Fatalf("Error initializing database: %v", err)
	}
	// handler closes the db for us
	defer os.Remove("test.db")

	handler, err := NewHandler("dhcpd0", config, db)
	if err != nil {
		t.Fatalf("Error initializing handler: %v", err)
	}
	defer handler.Close()

	if err != nil {
		t.Fatalf("Error configuring handler: %v", err)
	}

	go dhcp4.ListenAndServeIf("dhcpd0", handler)

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

func dumpInterfaces() {
	out, _ := exec.Command("ip", "addr").CombinedOutput()
	fmt.Println(string(out))
}

func testDHCP(t *testing.T) net.IP {
	cmd := exec.Command("dhclient", "-1", "-4", "-d", "-v", "dhclient0")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Error running dhclient: %v", err)
	}

	time.Sleep(time.Second)

	if err := exec.Command("pkill", "-KILL", "dhclient").Run(); err != nil {
		t.Fatalf("Error killing dhclient: %v", err)
	}

	cmd.Wait() // don't care

	dhc, err := netlink.LinkByName("dhclient0")
	if err != nil {
		t.Fatalf("Could not lookup dhclient veth pair: %v", err)
	}

	list, err := netlink.AddrList(dhc, netlink.FAMILY_V4)
	if err != nil {
		t.Fatalf("Could not list addresses for dhclient link: %v", err)
	}

	if len(list) != 1 {
		dumpInterfaces()
		t.Fatalf("Invalid addresses configured")
	}

	ip := list[0].IP

	if err := flushAddrs(dhc); err != nil {
		t.Fatalf("While flushing addresses: %v", err)
	}

	return ip
}

func flushAddrs(dhc netlink.Link) error {
	list, err := netlink.AddrList(dhc, netlink.FAMILY_V4)
	if err != nil {
		return errors.Wrap(err, "listing addresses")
	}

	for i := 0; i < len(list); i++ { // not using range because of pointer pass
		if err := netlink.AddrDel(dhc, &list[i]); err != nil {
			return errors.Wrapf(err, "deleting address %v", list[i].IP)
		}
	}

	return nil
}
