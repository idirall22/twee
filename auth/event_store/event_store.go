package notification

// EventStore interface.
type EventStore interface {
	Start() error
	Close() error
}
