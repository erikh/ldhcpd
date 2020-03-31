package dhcpd

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/db"
	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	defaultDBFile        = "ldhcpd.db"
	defaultLeaseDuration = 24 * time.Hour
)

// Range is for IP ranges
type Range struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func (r Range) String() string {
	return fmt.Sprintf("%v -> %v", r.From, r.To)
}

func (r Range) validate() error {
	from, to := r.Dimensions()
	if len(from) != 4 || len(to) != 4 {
		return errors.Errorf("invalid IP in range %v", r)
	}

	if dhcp4.IPLess(to, from) {
		return errors.Errorf("IPs are improperly specified in range: %v", r)
	}

	return nil
}

// Dimensions returns the IP addresses within the range
func (r Range) Dimensions() (net.IP, net.IP) {
	return net.ParseIP(r.From).To4(), net.ParseIP(r.To).To4()
}

// Config is the configuration of the dhcpd service
type Config struct {
	DNSServers    []string      `yaml:"dns_servers"`
	Gateway       string        `yaml:"gateway"`
	DBFile        string        `yaml:"db_file"`
	DynamicRange  Range         `yaml:"dynamic_range"`
	LeaseDuration time.Duration `yaml:"lease_duration"`
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

	return config, config.validateAndFix()
}

func (c *Config) validateAndFix() error {
	if err := c.DynamicRange.validate(); err != nil {
		return errors.Wrap(err, "could not validate dynamic range")
	}

	if len(c.GatewayIP()) != 4 {
		return errors.New("gateway IP is invalid")
	}

	if len(c.DNSServers) == 0 {
		c.DNSServers = []string{}
	}

	if len(c.DNSServers) > 0 && len(c.DNS()) == 0 {
		return errors.New("DNS servers contains invalid IPs")
	}

	if len(c.DNS())%4 != 0 {
		return errors.New("DNS servers contain invalid IPs")
	}

	if c.DBFile == "" {
		c.DBFile = defaultDBFile
	}

	if c.LeaseDuration == 0 {
		c.LeaseDuration = defaultLeaseDuration
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

// NewDB creates a new DB connection and migrates it if necessary.
func (c Config) NewDB() (*db.DB, error) {
	return db.NewDB(c.DBFile)
}
