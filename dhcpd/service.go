package dhcpd

import (
	"time"

	"github.com/krolaw/dhcp4"
	"github.com/sirupsen/logrus"
)

// ServeDHCP returns a dhcp response for a dhcp request.
func (h *Handler) ServeDHCP(req dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	switch msgType {
	case dhcp4.Discover:
		logrus.Infof("received discover")
		return dhcp4.ReplyPacket(req, dhcp4.Offer, h.ip, dhcp4.IPAdd(h.ip, 1), 24*time.Hour, h.options.SelectOrderOrAll(nil))
	case dhcp4.Request:
		logrus.Info("recieved request")
		return dhcp4.ReplyPacket(req, dhcp4.ACK, h.ip, req.CIAddr(), 24*time.Hour, h.options.SelectOrderOrAll(nil))
	case dhcp4.Release:
		logrus.Info("recieved release")
	case dhcp4.Decline:
		logrus.Info("recieved decline")
	}

	return nil
}
