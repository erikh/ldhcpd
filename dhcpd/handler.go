package dhcpd

import (
	"net"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/db"
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
		count, err := h.db.PurgeLeases()
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
func NewHandler(interfaceName, configFile string) (*Handler, error) {
	config, err := ParseConfig(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "while loading configuation")
	}

	return NewHandlerFromConfig(interfaceName, config)
}

// NewHandlerFromConfig is just like NewHandler only it accepts a config struct
// instead of a file.
func NewHandlerFromConfig(interfaceName string, config Config) (*Handler, error) {
	db, err := config.NewDB()
	if err != nil {
		return nil, errors.Wrap(err, "while bootstrapping dhcpd database")
	}

	intf, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, errors.Wrap(err, "error locating interface")
	}

	addrs, err := intf.Addrs()
	if err != nil {
		return nil, errors.Wrap(err, "error locating interface addresses")
	}

	if len(addrs) != 1 {
		return nil, errors.New("interface must be configured with exactly one address (for now)")
	}

	ip, ok := addrs[0].(*net.IPNet)
	if !ok {
		return nil, errors.New("internal error resolving interface address")
	}

	alloc, err := NewAllocator(db, config, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing allocator")
	}

	h := &Handler{
		ip:        ip.IP,
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
