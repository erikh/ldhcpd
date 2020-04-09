package proto

import (
	"context"
	"net"
	"time"

	"github.com/erikh/ldhcpd/db"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Handler is the control plane handler.
type Handler struct {
	db *db.DB
}

// Boot boots the grpc service
func Boot(db *db.DB) *grpc.Server {
	h := &Handler{db: db}

	s := grpc.NewServer()
	RegisterLeaseControlServer(s, h)

	return s
}

func toGRPC(lease *db.Lease) *Lease {
	return &Lease{
		MACAddress:    lease.MACAddress,
		IPAddress:     lease.IPAddress,
		Dynamic:       lease.Dynamic,
		Persistent:    lease.Persistent,
		LeaseEnd:      &timestamp.Timestamp{Seconds: lease.LeaseEnd.Unix()},
		LeaseGraceEnd: &timestamp.Timestamp{Seconds: lease.LeaseGraceEnd.Unix()},
	}
}

// SetLease creates an explicit lease with the parameters provided. It does not
// currently have any scoping rules other than it must be a valid lease in the
// networking sense. Whether or not the lease can be offered is another matter.
func (h *Handler) SetLease(ctx context.Context, lease *Lease) (*empty.Empty, error) {
	mac, err := net.ParseMAC(lease.MACAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "mac address is invalid: %v", err)
	}

	ip := net.ParseIP(lease.IPAddress)
	if len(ip) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "ip address is invalid")
	}

	if lease.LeaseEnd == nil || lease.LeaseGraceEnd == nil {
		return nil, status.Errorf(codes.InvalidArgument, "lease values are nil")
	}

	if err := h.db.SetLease(mac, ip.To4(), false, lease.Persistent, time.Unix(lease.LeaseEnd.Seconds, 0), time.Unix(lease.LeaseGraceEnd.Seconds, 0)); err != nil {
		return nil, status.Errorf(codes.Aborted, "failed to set lease: %v", err)
	}

	return &empty.Empty{}, nil
}

// GetLease retreives the lease for the mac address provided.
func (h *Handler) GetLease(ctx context.Context, mac *MACAddress) (*Lease, error) {
	m, err := net.ParseMAC(mac.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "mac address is invalid: %v", err)
	}

	lease, err := h.db.GetLease(m)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "could not retrieve lease: %v", err)
	}

	return toGRPC(lease), nil
}

// ListLeases lists all the leases we know about.
func (h *Handler) ListLeases(ctx context.Context, empty *empty.Empty) (*Leases, error) {
	list := []*Lease{}

	leases, err := h.db.ListLeases()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "could not list leases: %v", err)
	}

	for _, lease := range leases {
		list = append(list, toGRPC(lease))
	}

	return &Leases{List: list}, nil
}

// RemoveLease removes a lease.
func (h *Handler) RemoveLease(ctx context.Context, mac *MACAddress) (*empty.Empty, error) {
	m, err := net.ParseMAC(mac.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "mac address is invalid: %v", err)
	}

	if err := h.db.RemoveLease(m); err != nil {
		return nil, status.Errorf(codes.Aborted, "could not remove lease: %v", err)
	}

	return &empty.Empty{}, nil
}
