package dhcpd

import (
	"fmt"
	"time"

	"github.com/krolaw/dhcp4"
)

// ServeDHCP returns a dhcp response for a dhcp request.
func (h *Handler) ServeDHCP(req dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	switch msgType {
	case dhcp4.Discover:
		fmt.Println("recieved discover")
		fmt.Printf("%#v, %v\n", req, dhcp4.IPAdd(h.ip, 1))
		return dhcp4.ReplyPacket(req, dhcp4.Offer, h.ip, dhcp4.IPAdd(h.ip, 1), 24*time.Hour, h.options.SelectOrderOrAll(nil))
	case dhcp4.Request:
		fmt.Println("recieved request")
		return dhcp4.ReplyPacket(req, dhcp4.ACK, h.ip, req.CIAddr(), 24*time.Hour, h.options.SelectOrderOrAll(nil))
	case dhcp4.Release:
		fmt.Println("recieved release")
	case dhcp4.Decline:
		fmt.Println("recieved decline")
	}

	return nil
}
