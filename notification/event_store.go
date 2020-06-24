package notification

import "github.com/idirall22/twee/pb"

// EventStore interface.
type EventStore interface {
	Start() error
	Close() error
	Subscribe() <-chan *pb.Notification
}
