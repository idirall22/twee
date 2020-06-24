package timeline

import (
	"context"
	"fmt"

	eventstore "github.com/idirall22/twee/timeline/event_store"

	"github.com/idirall22/twee/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/idirall22/twee/pb"
	"github.com/idirall22/twee/timeline/store"
)

// Server timeline server service.
type Server struct {
	timelineStore      store.TimelineStore
	notificationClient *pb.NotificationServiceClient
	eventStore         eventstore.EventStore
	followClient       pb.FollowServiceClient
}

// NewTimelineServer create new timeline service.
func NewTimelineServer(
	s store.TimelineStore,
	es eventstore.EventStore,
	fc pb.FollowServiceClient,
) (*Server, error) {

	if s == nil {
		return nil, fmt.Errorf("Store should not be nil")
	}

	if fc == nil {
		return nil, fmt.Errorf("Follow service should not be nil")
	}

	// if es == nil {
	// 	return nil, fmt.Errorf("Event Store should not be NIL")
	// }

	// go es.Start()

	return &Server{
		timelineStore: s,
		eventStore:    es,
		followClient:  fc,
	}, nil
}

// Timeline user timeline home or self
func (s *Server) Timeline(req *pb.TimelineRequest, stream pb.TimelineService_TimelineServer) error {
	userID := req.UserId
	var followList []*pb.Follow

	if req.Type == pb.TimelineType_HOME {
		// get user infos from context
		userInfos, err := auth.GetUserInfosFromContext(stream.Context())
		if err != nil {
			return status.Errorf(codes.InvalidArgument, err.Error())
		}
		uc := stream.Context().Value(auth.ClaimKey("claims")).(*auth.UserClaims)
		userID = userInfos.ID

		ctx := metadata.AppendToOutgoingContext(context.Background(), auth.AuthKey, uc.Token)
		res, err := s.followClient.ListFollow(ctx, &pb.RequestListFollow{
			Follower:   req.UserId,
			FollowType: pb.FollowListType_FOLLOWER,
		})

		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		followList = res.Follows
	}

	err := s.timelineStore.List(
		stream.Context(),
		userID,
		followList,
		req.Type,
		func(tweet *pb.Tweet) error {
			err := stream.Send(&pb.TimelineResponse{Tweet: tweet})
			if err != nil {
				return status.Errorf(codes.Internal, "Could not send tweet: %v", err)
			}
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "Could not list timeline: %v", err)
	}

	return nil
}
