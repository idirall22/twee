package teventstore

import (
	"context"
	"fmt"

	"github.com/idirall22/twee/common"
	"github.com/idirall22/twee/pb"
	"github.com/nats-io/stan.go"
)

// NatsStreamingEventStore event store.
type NatsStreamingEventStore struct {
	cc            stan.Conn
	subject       string
	notifications chan *pb.NewNotification
}

// NewNatsStreamingEventStore create new stan event store.
func NewNatsStreamingEventStore(subject string) (*NatsStreamingEventStore, error) {
	sc, err := stan.Connect("test-cluster", "test-cluster")
	if err != nil {
		return nil, fmt.Errorf("Could not connect to nats streaming server: %v", err)
	}

	return &NatsStreamingEventStore{
		cc:            sc,
		subject:       subject,
		notifications: make(chan *pb.NewNotification, 128),
	}, nil
}

// Start event store.
func (e *NatsStreamingEventStore) Start() error {
	for {
		select {
		case n := <-e.notifications:
			data, err := common.ProtobufToJSON(n)
			if err != nil {
				return fmt.Errorf("Could not serialize data to json: %v", err)
			}

			err = e.cc.Publish(e.subject, []byte(data))
			if err != nil {
				return fmt.Errorf("Could not publish to nats: %v", err)
			}
		}
	}
}

// Publish to event store
func (e *NatsStreamingEventStore) Publish(ctx context.Context, n *pb.NewNotification) error {
	e.notifications <- n
	return nil
}

// Close event store connection.
func (e *NatsStreamingEventStore) Close() error {
	return e.cc.Close()
}
