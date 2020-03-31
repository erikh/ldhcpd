package dhcpd

import (
	"net"

	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
)

// Handler is the dhpcd handler for serving requests.
type Handler struct {
	ip      net.IP
	options dhcp4.Options
	config  Config
}

// NewHandler creates a new dhcpd handler.
func NewHandler(interfaceName, configFile string) (*Handler, error) {
	config, err := ParseConfig(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "while loading configuation")
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

	return &Handler{
		ip:     ip.IP,
		config: config,
		options: dhcp4.Options{
			dhcp4.OptionSubnetMask:       ip.Mask,
			dhcp4.OptionRouter:           config.GatewayIP(),
			dhcp4.OptionDomainNameServer: config.DNS(),
		},
	}, nil
}