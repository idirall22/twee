package follow

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/idirall22/twee/auth"
	eventstore "github.com/idirall22/twee/follow/event_store"
	"github.com/idirall22/twee/follow/store"
	"github.com/idirall22/twee/pb"
)

// Server follow service struct.
type Server struct {
	followStore        store.FollowStore
	notificationClient *pb.NotificationServiceClient
	eventStore         eventstore.EventStore
}

// NewFollowServer create new follow service.
func NewFollowServer(s store.FollowStore, es eventstore.EventStore) (*Server, error) {
	if s == nil {
		return nil, fmt.Errorf("Store should not be NIL")
	}

	// if es == nil {
	// 	return nil, fmt.Errorf("Event Store should not be NIL")
	// }

	// go es.Start()

	return &Server{
		followStore: s,
		eventStore:  es,
	}, nil
}

// ToggleFollow a user
func (s *Server) ToggleFollow(ctx context.Context, req *pb.RequestFollow) (*pb.ResponseFollow, error) {
	// get user infos from context
	userInfos, err := auth.GetUserInfosFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	followee := req.Followee
	follower := userInfos.ID
	_, err = s.followStore.ToggleFollow(ctx, follower, followee)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not toggle follow: %v", err)
	}

	// publish follow event.
	// go s.eventStore.Publish(ctx, &pb.FollowEvent{
	// 	Action:   action,
	// 	Followee: followee,
	// 	Follower: follower,
	// })

	return &pb.ResponseFollow{}, nil
}

// ListFollow list followee or followers
func (s *Server) ListFollow(ctx context.Context, req *pb.RequestListFollow) (*pb.ResponseListFollow, error) {
	followsList, err := s.followStore.ListFollow(
		ctx,
		req.Follower,
		req.Followee,
		req.FollowType,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error to list follow: %v", err)
	}
	return &pb.ResponseListFollow{Follows: followsList}, nil
}
