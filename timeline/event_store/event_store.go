package eventstore

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// EventStore interface
type EventStore interface {
	// Start event store.
	Start() error
	// Publish to event store
	Publish(ctx context.Context, n *pb.TweetEvent) error
	// Close event store connection.
	Close() error
}
