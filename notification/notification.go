package notification

import (
	"fmt"

	"google.golang.org/grpc/codes"

	neventstore "github.com/idirall22/twee/notification/event_store/stan"
	"google.golang.org/grpc/status"

	nstore "github.com/idirall22/twee/notification/store/postgres"
	"github.com/idirall22/twee/pb"

	option "github.com/idirall22/twee/options"
)

// Server notification service server.
type Server struct {
	eventStore EventStore
}

// NewNotificationServer create notification server service
func NewNotificationServer(opts *option.PostgresOptions) (*Server, error) {
	nStore, err := nstore.NewPostgresNotificationStore(opts)

	if err != nil {
		return nil, fmt.Errorf("Could not Start store: %v", err)
	}

	es, err := neventstore.NewNatsStreamingEventStore("tweets")
	if err != nil {
		return nil, fmt.Errorf("Could not Start event store: %v", err)
	}

	return &Server{
		notificationStore: nStore,
		eventStore:        es,
	}, nil
}

// Notify client
func (s *Server) Notify(req *pb.NotifyRequest, stream pb.NotificationService_NotifyServer) error {
	for {
		err := stream.Send(&pb.NotifyResponse{})
		if err != nil {
			return status.Errorf(codes.Internal, "Error to send notification: %v", err)
		}
	}
}
