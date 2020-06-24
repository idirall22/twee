package feventstore

import (
	"context"
	"fmt"

	"github.com/idirall22/twee/common"
	"github.com/idirall22/twee/pb"
	"github.com/nats-io/stan.go"
)

// NatsStreamingEventStore event store.
type NatsStreamingEventStore struct {
	cc           stan.Conn
	subject      string
	followEvents chan *pb.FollowEvent
}

// NewNatsStreamingEventStore create new stan event store.
func NewNatsStreamingEventStore(subject, clusterID, clientID string, option ...stan.Option) (*NatsStreamingEventStore, error) {
	sc, err := stan.Connect(clusterID, clientID, option...)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to nats streaming server: %v", err)
	}

	return &NatsStreamingEventStore{
		cc:           sc,
		subject:      subject,
		followEvents: make(chan *pb.FollowEvent, 128),
	}, nil
}

// Start event store.
func (e *NatsStreamingEventStore) Start() error {
	for {
		select {
		case n := <-e.followEvents:
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
func (e *NatsStreamingEventStore) Publish(ctx context.Context, te *pb.FollowEvent) error {
	e.followEvents <- te
	return nil
}

// Close event store connection.
func (e *NatsStreamingEventStore) Close() error {
	return e.cc.Close()
}
