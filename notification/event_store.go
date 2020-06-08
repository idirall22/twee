package notification

// EventStore interface.
type EventStore interface {
	Start(subject string) error
	Close() error
}
