package timeline

import (
	"fmt"
	"time"

	"github.com/idirall22/twee/auth"
	tlstore "github.com/idirall22/twee/timeline/store/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
)

// Server timeline server service
type Server struct {
	timelineStore Store
}

// NewTimelineServer create new timeline server
func NewTimelineServer() (*Server, error) {
	opts := option.NewPostgresOptions(
		"0.0.0.0",
		"postgres",
		"password",
		"twee",
		3,
		5432,
		time.Second,
	)

	tlStore, err := tlstore.NewPostgresTimelineStore(opts)

	if err != nil {
		return nil, fmt.Errorf("Could not Start store: %v", err)
	}

	return &Server{
		timelineStore: tlStore,
	}, nil
}

// Timeline user timeline home or self
func (s *Server) Timeline(req *pb.TimelineRequest, stream pb.TimelineService_TimelineServer) error {
	// get user infos from context
	userInfos, err := auth.GetUserInfosFromContext(stream.Context())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.timelineStore.List(
		stream.Context(),
		userInfos.ID,
		req.Type,
		func(tweet *pb.Tweet) error {
			err = stream.Send(&pb.TimelineResponse{Tweet: tweet})
			if err != nil {
				return status.Errorf(codes.Internal, "Could not send tweet: %v", err)
			}
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "")
	}

	return nil
}
