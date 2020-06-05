package follow

import (
	"context"
	"fmt"
	"time"

	"github.com/idirall22/twee/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fstore "github.com/idirall22/twee/follow/store/postgres"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
)

// Server follow service struct.
type Server struct {
	followStore *fstore.PostgresFollowStore
}

// NewFollowServer create new follow service.
func NewFollowServer() (*Server, error) {
	opts := option.NewPostgresOptions(
		"0.0.0.0",
		"postgres",
		"password",
		"twee",
		3,
		5432,
		time.Second,
	)

	followStore, err := fstore.NewPostgresFollowStore(opts)

	if err != nil {
		return nil, fmt.Errorf("Could not Start store: %v", err)
	}
	return &Server{
		followStore: followStore,
	}, nil
}

// ToggleFollow a user
func (s *Server) ToggleFollow(ctx context.Context, req *pb.RequestFollow) (*pb.ResponseFollow, error) {
	// get user infos from context
	userInfos, err := auth.GetUserInfosFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.followStore.ToggleFollow(ctx, userInfos.ID, req.Followee)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not Follow: %v", err)
	}

	return &pb.ResponseFollow{}, nil
}
