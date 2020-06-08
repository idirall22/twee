package notification

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store store notification interface.
type Store interface {
	// New create notification
	New(ctx context.Context, notif *pb.Notification) error
	// List notifications
	List(ctx context.Context, userID int64, found func(n *pb.Notification) error) error
}
