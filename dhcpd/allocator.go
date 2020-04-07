package dhcpd

import (
	"net"
	"sync"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/db"
	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
)

// ErrRangeExhausted is returned when the IP range is exhausted
var ErrRangeExhausted = errors.New("IP range exhausted")

// Allocator allocates IP addresses from a range
type Allocator struct {
	config Config
	db     *db.DB

	lastIP      net.IP
	lastIPMutex sync.Mutex
}

// NewAllocator creates a new allocator
func NewAllocator(db *db.DB, c Config, initial net.IP) (*Allocator, error) {
	if initial == nil {
		initial = net.ParseIP(c.DynamicRange.From)
	}

	return &Allocator{
		config: c,
		db:     db,
		lastIP: dhcp4.IPAdd(initial, -1),
	}, nil
}

// Allocate or Retrieve an IP address for a mac. renew states that if there is
// already an IP present in the leases table for this mac, to renew the lease
// if necessary.
func (a *Allocator) Allocate(mac net.HardwareAddr, renew bool) (net.IP, error) {
	now := time.Now()
	// FIXME returning lease end here may help with some distributed race conditions we're seeing
	l, err := a.db.GetLease(mac)
	if err == nil {
		if (renew && (l.LeaseEnd.Before(now) || l.LeaseGraceEnd.Before(now))) || l.Persistent {
			leaseEnd := now.Add(a.config.Lease.Duration)
			l, err = a.db.RenewLease(mac, leaseEnd, leaseEnd.Add(a.config.Lease.GracePeriod))
			if err != nil {
				return nil, errors.Wrapf(err, "could not renew lease for mac [%v] ip [%v]", mac, a.lastIP)
			}
		}

		return l.IP(), nil
	}
	first, last := a.config.DynamicRange.Dimensions()

	a.lastIPMutex.Lock()
	defer a.lastIPMutex.Unlock()

	// calculate these ahead of time to save a few cycles
	leaseEnd := now.Add(a.config.Lease.Duration)
	gracePeriodEnd := leaseEnd.Add(a.config.Lease.GracePeriod)

	var foundFirst, foundFirstClearedGrace bool
	for {
		ip := dhcp4.IPAdd(a.lastIP, 1)

		if !dhcp4.IPInRange(first, last, ip) {
			if foundFirst {
				if foundFirstClearedGrace {
					return nil, ErrRangeExhausted
				}

				_, err := a.db.PurgeLeases(true)
				if err != nil {
					return nil, errors.Wrap(err, "trying to clean up lease table")
				}

				foundFirstClearedGrace = true
			}
			a.lastIP = first
			foundFirst = true
		} else {
			a.lastIP = ip
		}

		if err := a.db.SetLease(mac, a.lastIP, true, false, leaseEnd, gracePeriodEnd); err != nil {
			continue
		}

		return a.lastIP, nil
	}
}
