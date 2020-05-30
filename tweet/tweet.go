package tweet

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/idirall22/twee/pb"
	postgresstore "github.com/idirall22/twee/tweet/store/postgres"
	"google.golang.org/grpc/status"
)

// Server server
type Server struct {
	tweetStore Store
}

// NewServer create new tweet server
func NewServer() (*Server, error) {
	opts := postgresstore.NewPostgresOptions(
		"localhost",
		"postgres",
		"password",
		"tweets",
		3,
		5432,
		time.Second,
	)

	pStore, err := postgresstore.NewPostgresTweetStore(opts)

	if err != nil {
		return nil, fmt.Errorf("Could not Start store: %v", err)
	}

	return &Server{
		tweetStore: pStore,
	}, nil
}

// Create a tweet.
func (s *Server) Create(ctx context.Context, req *pb.CreateTweetRequest) (*pb.CreateTweetResponse, error) {
	content := req.Content
	if len(content) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "tweet Content is empty")
	}

	id, err := s.tweetStore.Create(ctx, content)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error to create a tweet: %v", err)
	}

	res := &pb.CreateTweetResponse{
		Id: id,
	}
	return res, nil
}

// Update a tweet.
func (s *Server) Update(ctx context.Context, req *pb.UpdateTweetRequest) (*pb.UpdateTweetResponse, error) {
	return nil, nil
}

// Delete a tweet.
func (s *Server) Delete(ctx context.Context, req *pb.DeleteTweetRequest) (*pb.DeleteTweetResponse, error) {
	return nil, nil
}

// Get tweet by user id and tweet id.
func (s *Server) Get(ctx context.Context, req *pb.GetTweetRequest) (*pb.GetTweetResponse, error) {
	return nil, nil
}

// List a user tweets using user id.
func (s *Server) List(req *pb.ListTweetRequest, stream pb.TweetService_ListServer) error {
	return nil
}

// Close store connection
func (s *Server) Close() error {
	return nil
}
