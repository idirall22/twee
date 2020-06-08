package tweet

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// EventStore interface
type EventStore interface {
	// Publish to event store
	Publish(ctx context.Context, n *pb.NewNotification) error
	// Close event store connection.
	Close() error
}
