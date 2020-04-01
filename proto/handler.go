package proto

import (
	context "context"

	"code.hollensbe.org/erikh/ldhcpd/db"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// Handler is the control plane handler.
type Handler struct {
	db *db.DB
}

// NewHandler is a handler for the grpc control plane.
func NewHandler(db *db.DB) *Handler {
	return &Handler{db: db}
}

// SetLease creates an explicit lease with the parameters provided. It does not
// currently have any scoping rules other than it must be a valid lease in the
// networking sense. Whether or not the lease can be offered is another matter.
func (h *Handler) SetLease(ctx context.Context, lease *Lease) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

// GetLease retreives the lease for the mac address provided.
func (h *Handler) GetLease(ctx context.Context, mac *MACAddress) (*Lease, error) {
	return &Lease{}, nil
}

// ListLeases lists all the leases we know about.
func (h *Handler) ListLeases(ctx context.Context, empty *empty.Empty) (*Leases, error) {
	return &Leases{List: []*Lease{}}, nil
}
