package notification

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nstore "github.com/idirall22/twee/notification/store/postgres"
	"github.com/idirall22/twee/pb"

	option "github.com/idirall22/twee/options"
)

// Server notification service server.
type Server struct {
	notificationStore Store
	notifications     chan *pb.Notification
}

// NewNotificationServer create notification server service
func NewNotificationServer(opts *option.PostgresOptions) (*Server, error) {
	nStore, err := nstore.NewPostgresNotificationStore(opts)

	if err != nil {
		return nil, fmt.Errorf("Could not Start store: %v", err)
	}
	return &Server{
		notificationStore: nStore,
		notifications:     make(chan *pb.Notification, 1000),
	}, nil
}

// Notify client
func (s *Server) Notify(req *pb.NotifyRequest, stream pb.NotificationService_NotifyServer) error {
	return nil
}

// New create new notification
func (s *Server) New(ctx context.Context, req *pb.NewNotitficationRequest) (*pb.NewNotitficationResponse, error) {
	err := s.notificationStore.New(ctx, req.Notification)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error to create notification: %v", err)
	}
	return &pb.NewNotitficationResponse{}, nil
}
