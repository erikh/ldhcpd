package dhcpd

import (
	"github.com/krolaw/dhcp4"
	"github.com/sirupsen/logrus"
)

// ServeDHCP returns a dhcp response for a dhcp request.
func (h *Handler) ServeDHCP(req dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	switch msgType {
	case dhcp4.Discover:
		logrus.Infof("received discover from %v", req.CHAddr())

		ip, err := h.allocator.Allocate(req.CHAddr(), true)
		if err != nil {
			logrus.Errorf("Error allocating IP for %v: %v", req.CHAddr(), err)
			return nil
		}

		return dhcp4.ReplyPacket(req, dhcp4.Offer, h.ip, ip, h.config.LeaseDuration, h.options.SelectOrderOrAll(nil))
	case dhcp4.Request:
		logrus.Infof("received request for %v from %v", req.CIAddr(), req.CHAddr())

		ip, err := h.allocator.Allocate(req.CHAddr(), true)
		if err != nil {
			logrus.Errorf("Error allocating IP for %v: %v", req.CHAddr(), err)
			return nil
		}

		if req.CIAddr().IsUnspecified() || req.CIAddr().Equal(ip) {
			return dhcp4.ReplyPacket(req, dhcp4.ACK, h.ip, ip, h.config.LeaseDuration, h.options.SelectOrderOrAll(nil))
		}

		return nil
	case dhcp4.Release:
		logrus.Info("received release")
	case dhcp4.Decline:
		logrus.Info("received decline")
	}

	return nil
}
