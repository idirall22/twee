package notification

import (
	"fmt"

	neventstore "github.com/idirall22/twee/notification/event_store/stan"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/idirall22/twee/pb"

	option "github.com/idirall22/twee/options"
)

// Server notification service server.
type Server struct {
	eventStore  EventStore
	connections map[int64]*pb.NotificationService_NotifyServer
}

// NewNotificationServer create notification server service
func NewNotificationServer(opts *option.PostgresOptions) (*Server, error) {
	es, err := neventstore.NewNatsStreamingEventStore("tweets", opts)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to event store: %v", err)
	}
	go es.Start()

	return &Server{
		eventStore:  es,
		connections: make(map[int64]*pb.NotificationService_NotifyServer, 128),
	}, nil
}

// Notify client
func (s *Server) Notify(req *pb.NotifyRequest, stream pb.NotificationService_NotifyServer) error {
	notificationsChan := s.eventStore.Subscribe()
	// 1- get user id
	// 2- store user stream in connections

	for {
		select {
		case n := <-notificationsChan:
			fmt.Println(n)
			err := stream.Send(&pb.NotifyResponse{Notification: n})
			if err != nil {
				return status.Errorf(codes.Internal, "Error to send notification: %v", err)
			}
		}
	}
}

// Close service.
func (s *Server) Close() error {
	return s.eventStore.Close()
}
