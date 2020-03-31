package dhcpd

import (
	"io/ioutil"
	"net"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config is the configuration of the dhcpd service
type Config struct {
	DNSServers []string `yaml:"dns_servers"`
	Gateway    string   `yaml:"gateway"`
}

// ParseConfig parses the configuration in the file and returns it.
func ParseConfig(filename string) (Config, error) {
	var config Config

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, errors.Wrap(err, "while reading configuration file")
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return config, errors.Wrap(err, "error while parsing configuration file")
	}

	return config, config.validate()
}

func (c *Config) validate() error {
	if len(c.GatewayIP()) != 4 {
		return errors.New("gateway IP is invalid")
	}

	if len(c.DNS())%4 != 0 {
		return errors.New("DNS servers contain invalid IPs")
	}

	return nil
}

// GatewayIP returns the gateway IP
func (c Config) GatewayIP() net.IP {
	return net.ParseIP(c.Gateway).To4()
}

// DNS returns the IP addresses associated with the DNS servers.
func (c Config) DNS() []byte {
	ips := []byte{}
	for _, srv := range c.DNSServers {
		ips = append(ips, net.ParseIP(srv).To4()...)
	}

	return ips
}
