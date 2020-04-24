package dhcpd

import (
	"net"
	"time"

	"github.com/erikh/ldhcpd/db"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type dhcpOptions map[dhcpv4.OptionCode]dhcpv4.OptionValue

// Handler is the dhpcd handler for serving requests.
type Handler struct {
	ip        net.IP
	options   dhcpOptions
	config    Config
	db        *db.DB
	allocator *Allocator
	closed    bool
}

func (h *Handler) purgeLeases() {
	for {
		time.Sleep(time.Second)
		count, err := h.db.PurgeLeases(false)
		if err != nil {
			logrus.Errorf("While purging leases: %v", err)
			continue
		}

		if count != 0 {
			logrus.Infof("Periodic purge of %d expired leases occurred", count)
		}
	}
}

// InterfaceIP gets the most likely interface IP safe for listening on DHCP.
func InterfaceIP(interfaceName string) (*net.IPNet, error) {
	intf, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, errors.Wrap(err, "error locating interface")
	}

	addrs, err := intf.Addrs()
	if err != nil {
		return nil, errors.Wrap(err, "error locating interface addresses")
	}

	var ip *net.IPNet

	for _, addr := range addrs {
		var ok bool
		ip, ok = addr.(*net.IPNet)
		if !ok {
			return nil, errors.New("internal error resolving interface address")
		}

		if ip.IP.IsGlobalUnicast() {
			logrus.Infof("Selecting %v to serve DHCP", ip.IP.String())
			break
		}
	}

	if ip == nil {
		return nil, errors.Errorf("Could not find a suitable IP for serving on interface %v", interfaceName)
	}

	return ip, nil
}

// NewHandler creates a new dhcpd handler.
func NewHandler(ip *net.IPNet, config Config, db *db.DB) (*Handler, error) {

	alloc, err := NewAllocator(db, config, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing allocator")
	}

	h := &Handler{
		ip:        ip.IP.To4(),
		config:    config,
		db:        db,
		allocator: alloc,
		options: dhcpOptions{
			dhcpv4.OptionSubnetMask:       dhcpv4.IP(ip.Mask),
			dhcpv4.OptionRouter:           dhcpv4.IP(config.GatewayIP()),
			dhcpv4.OptionDomainNameServer: dhcpv4.IPs(config.DNS()),
		},
	}

	// FIXME this should be a toggle
	go h.purgeLeases()

	return h, nil
}

// Close the handler
func (h *Handler) Close() error {
	h.closed = true
	return h.db.Close()
}
