package notification

import (
	"errors"
	"fmt"

	"github.com/idirall22/twee/follow"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notifeventstore "github.com/idirall22/twee/notification/event_store"
	"github.com/idirall22/twee/notification/store"
	"github.com/idirall22/twee/pb"
)

// Server notification service server.
type Server struct {
	running     bool
	store       store.Store
	eventStore  notifeventstore.EventStore
	connections map[int64]*pb.NotificationService_NotifyServer
}

// NewNotificationServer create notification server service
func NewNotificationServer(
	ns store.Store,
	es notifeventstore.EventStore,
	fs *follow.Server,
) (*Server, error) {
	if ns == nil {
		return nil, errors.New("Notification store should not be nil")
	}

	if es == nil {
		return nil, errors.New("Notification event store should not be nil")
	}

	if fs == nil {
		return nil, errors.New("Follow service should not be nil")
	}

	return &Server{
		store:       ns,
		eventStore:  es,
		connections: make(map[int64]*pb.NotificationService_NotifyServer, 128),
	}, nil
}

// Start notification service
func (s *Server) Start() error {
	if !s.running {
		s.running = true
		return s.eventStore.Start()
	}
	return errors.New("Serivce already started")
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
