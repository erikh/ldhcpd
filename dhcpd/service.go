package dhcpd

import (
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/sirupsen/logrus"
)

func (h *Handler) configureReply(m *dhcpv4.DHCPv4, mt dhcpv4.MessageType) (*dhcpv4.DHCPv4, error) {
	rep, err := dhcpv4.NewReplyFromRequest(m)
	if err != nil {
		return nil, err
	}

	rep.UpdateOption(dhcpv4.OptMessageType(mt))
	rep.UpdateOption(dhcpv4.OptServerIdentifier(h.ip))
	rep.UpdateOption(dhcpv4.OptIPAddressLeaseTime(h.config.Lease.Duration))

	for opt, val := range h.options {
		rep.UpdateOption(dhcpv4.Option{Code: opt, Value: val})
	}

	return rep, nil
}

// ServeDHCP returns a dhcp response for a dhcp request.
func (h *Handler) ServeDHCP(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	if h.closed {
		return
	}

	switch m.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		logrus.Infof("received discover from %v", m.ClientHWAddr)

		ip, err := h.allocator.Allocate(m.ClientHWAddr, true, nil)
		if err != nil {
			logrus.Errorf("Error allocating IP for %v: %v", m.ClientHWAddr, err)
			return
		}

		logrus.Infof("Generated lease for mac [%v] ip [%v]", m.ClientHWAddr, ip)

		rep, err := h.configureReply(m, dhcpv4.MessageTypeOffer)
		if err != nil {
			logrus.Errorf("While configuring discover reply: %v", err)
			return
		}

		rep.YourIPAddr = ip

		if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
			logrus.Errorf("Error replying to DHCP discover: %v", err)
			return
		}
	case dhcpv4.MessageTypeRequest:
		logrus.Infof("received request for %v from %v", m.ClientHWAddr, m.ClientHWAddr)

		preferredIP := net.IP(m.Options[uint8(dhcpv4.OptionRequestedIPAddress)])
		if preferredIP == nil {
			preferredIP = m.ClientIPAddr
		}

		ip, err := h.allocator.Allocate(m.ClientHWAddr, true, preferredIP)
		if err != nil {
			logrus.Errorf("Error allocating IP for %v: %v", m.ClientHWAddr, err)
			// FIXME NAK here
			return
		}

		logrus.Infof("Lease obtained for mac [%v] ip [%v]", m.ClientHWAddr, ip)

		rep, err := h.configureReply(m, dhcpv4.MessageTypeAck)
		if err != nil {
			logrus.Errorf("While configuring discover reply: %v", err)
			return
		}

		rep.YourIPAddr = ip

		if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
			logrus.Errorf("Error replying to DHCP request: %v", err)
			return
		}
	case dhcpv4.MessageTypeRelease:
		logrus.Info("received release")
	case dhcpv4.MessageTypeDecline:
		logrus.Info("received decline")
	}
}
