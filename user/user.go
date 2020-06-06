package user

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/idirall22/twee/pb"
	ustore "github.com/idirall22/twee/user/store"

	option "github.com/idirall22/twee/options"
)

// Server user service server
type Server struct {
	userStore Store
}

// NewUserServer create user server service
func NewUserServer(opts *option.PostgresOptions) (*Server, error) {
	uStore, err := ustore.NewPostgresUserStore(opts)

	if err != nil {
		return nil, fmt.Errorf("Could not Start store: %v", err)
	}
	return &Server{
		userStore: uStore,
	}, nil
}

// List users
func (s *Server) List(req *pb.RequestListUsers, stream pb.UserService_ListServer) error {
	err := s.userStore.List(
		stream.Context(),
		req.GetLimit(),
		req.GetOffset(),
		func(user *pb.User) error {
			res := &pb.ResposneUser{User: user}
			err := stream.Send(res)
			if err != nil {
				return status.Errorf(codes.Internal, "Error stream profiles: %v", err)
			}
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "Error list profiles: %v", err)
	}

	return nil
}

// Profile get user profile
func (s *Server) Profile(ctx context.Context, req *pb.RequestUserProfile) (*pb.ResposneUser, error) {
	if len(req.GetUsername()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Username should not be empty")
	}

	user, err := s.userStore.Profile(ctx, req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error to fetch user profile: %v", err)
	}

	return &pb.ResposneUser{User: user}, nil
}
