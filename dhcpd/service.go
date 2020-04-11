package dhcpd

import (
	"net"

	"github.com/krolaw/dhcp4"
	"github.com/sirupsen/logrus"
)

// ServeDHCP returns a dhcp response for a dhcp request.
func (h *Handler) ServeDHCP(req dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	switch msgType {
	case dhcp4.Discover:
		logrus.Infof("received discover from %v", req.CHAddr())

		ip, err := h.allocator.Allocate(req.CHAddr(), true, nil)
		if err != nil {
			logrus.Errorf("Error allocating IP for %v: %v", req.CHAddr(), err)
			return dhcp4.ReplyPacket(req, dhcp4.NAK, h.ip, req.CIAddr(), 0, nil)
		}

		logrus.Infof("Generated lease for mac [%v] ip [%v]", req.CHAddr(), ip)

		return dhcp4.ReplyPacket(req, dhcp4.Offer, h.ip, ip, h.config.Lease.Duration, h.options.SelectOrderOrAll(nil))
	case dhcp4.Request:
		logrus.Infof("received request for %v from %v", req.CIAddr(), req.CHAddr())

		preferredIP := net.IP(options[dhcp4.OptionRequestedIPAddress])
		if preferredIP == nil {
			preferredIP = req.CIAddr()
		}

		ip, err := h.allocator.Allocate(req.CHAddr(), true, preferredIP)
		if err != nil {
			logrus.Errorf("Error allocating IP for %v: %v", req.CHAddr(), err)
			return dhcp4.ReplyPacket(req, dhcp4.NAK, h.ip, req.CIAddr(), 0, nil)
		}

		logrus.Infof("Lease obtained for mac [%v] ip [%v]", req.CHAddr(), ip)

		return dhcp4.ReplyPacket(req, dhcp4.ACK, h.ip, ip, h.config.Lease.Duration, h.options.SelectOrderOrAll(nil))
	case dhcp4.Release:
		logrus.Info("received release")
	case dhcp4.Decline:
		logrus.Info("received decline")
	}

	return nil
}
