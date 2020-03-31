package dhcpd

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

const defaultBridge = "testbridge0"

var initialInterfaces = []string{"dhcpd0", "dhclient0"}

func addVethPair(name string, bridge *netlink.Bridge) error {
	peerName := name + "-peer"

	la := netlink.NewLinkAttrs()
	la.Name = name
	if err := netlink.LinkAdd(&netlink.Veth{LinkAttrs: la, PeerName: peerName}); err != nil {
		return errors.Wrap(err, "could not create veth pair")
	}

	peer, err := netlink.LinkByName(peerName)
	if err != nil {
		return errors.Wrap(err, "could not locate created peer")
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

	for _, name := range initialInterfaces {
		if err := addVethPair(name, bridge); err != nil {
			t.Fatalf("Could not add veth pair: %v", err)
		}

		link, err := netlink.LinkByName(name)
		if err != nil {
			t.Fatalf("Could not find newly added link %v: %v", name, err)
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
	for _, name := range initialInterfaces {
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
}
