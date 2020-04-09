package dhcpd

import (
	"net"
	"time"

	"github.com/erikh/ldhcpd/db"
	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Handler is the dhpcd handler for serving requests.
type Handler struct {
	ip        net.IP
	options   dhcp4.Options
	config    Config
	db        *db.DB
	allocator *Allocator
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

// NewHandler creates a new dhcpd handler.
func NewHandler(interfaceName string, config Config, db *db.DB) (*Handler, error) {
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

	alloc, err := NewAllocator(db, config, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing allocator")
	}

	h := &Handler{
		ip:        ip.IP.To4(),
		config:    config,
		db:        db,
		allocator: alloc,
		options: dhcp4.Options{
			dhcp4.OptionSubnetMask:       ip.Mask,
			dhcp4.OptionRouter:           config.GatewayIP(),
			dhcp4.OptionDomainNameServer: config.DNS(),
		},
	}

	// FIXME this should be a toggle
	go h.purgeLeases()

	return h, nil
}

// Close the handler
func (h *Handler) Close() error {
	return h.db.Close()
}
