package neventstore

import (
	"context"
	"fmt"
	"time"

	"github.com/idirall22/twee/pb"

	"github.com/idirall22/twee/common"

	nstore "github.com/idirall22/twee/notification/store/postgres"
	option "github.com/idirall22/twee/options"

	"github.com/idirall22/twee/notification/store"
	"github.com/nats-io/stan.go"
)

// NatsStreamingEventStore struct.
type NatsStreamingEventStore struct {
	notificationStore store.Store
	cc                stan.Conn
	subject           string
	newNotifications  chan string
	notifications     chan *pb.Notification
	done              chan error
}

// NewNatsStreamingEventStore create new NatsStreamingEventStore.
func NewNatsStreamingEventStore(subject string, opts *option.PostgresOptions) (*NatsStreamingEventStore, error) {
	cc, err := stan.Connect("test-cluster", "test-cluster-01")
	if err != nil {
		return nil, fmt.Errorf("Could not connect to nats streaming server: %v", err)
	}

	ns, err := nstore.NewPostgresNotificationStore(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not create notification store: %v", err)
	}
	return &NatsStreamingEventStore{
		notificationStore: ns,
		subject:           subject,
		cc:                cc,
		newNotifications:  make(chan string, 128),
	}, nil
}

// Start NatsStreamingEventStore.
func (e *NatsStreamingEventStore) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	go func() {
		for {
			select {
			case <-e.done:
				return
			case msg := <-e.newNotifications:
				nn := &pb.NewNotification{}
				err := common.JSONToProtobufMessage(msg, nn)
				if err != nil {
					e.done <- fmt.Errorf("Could not parse json: %v", err)
					break
				}

				err = e.notificationStore.New(ctx, nn, e.notifications)
				if err != nil {
					e.done <- fmt.Errorf("Could not create notifications: %v", err)
					break
				}
			}
		}
	}()

	sub, err := e.cc.Subscribe(e.subject, func(msg *stan.Msg) {
		e.newNotifications <- string(msg.Data)
	}, stan.DeliverAllAvailable())

	if err != nil {
		return fmt.Errorf("Could not subscribe to nats: %v", err)
	}

	defer sub.Close()

	return <-e.done
}

// Close close
func (e *NatsStreamingEventStore) Close() error {
	return e.cc.Close()
}

// Subscribe get notification channel
func (e *NatsStreamingEventStore) Subscribe() <-chan *pb.Notification {
	return e.notifications
}
