package tweet

import (
	"context"
	"fmt"
	"time"

	// postgres driver
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
	postgresstore "github.com/idirall22/twee/tweet/store/postgres"
)

// Server server
type Server struct {
	tweetStore Store
}

// NewServer create new tweet server
func NewServer() (*Server, error) {
	opts := option.NewPostgresOptions(
		"0.0.0.0",
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
	id := req.GetId()
	content := req.GetContent()

	if id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid id")
	}

	if len(content) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "empty content")
	}

	err := s.tweetStore.Update(ctx, id, content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error to update the tweet: %v", err)
	}
	return nil, nil
}

// Delete a tweet.
func (s *Server) Delete(ctx context.Context, req *pb.DeleteTweetRequest) (*pb.DeleteTweetResponse, error) {
	id := req.GetId()

	err := s.tweetStore.Delete(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not delete tweet: %v", err)
	}

	return nil, nil
}

// Get tweet by user id and tweet id.
func (s *Server) Get(ctx context.Context, req *pb.GetTweetRequest) (*pb.GetTweetResponse, error) {
	id := req.GetId()
	tweet, err := s.tweetStore.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get tweet: %v", err)
	}
	res := &pb.GetTweetResponse{
		Tweet: tweet,
	}
	return res, nil
}

// List a user tweets using user id.
func (s *Server) List(req *pb.ListTweetRequest, stream pb.TweetService_ListServer) error {
	// TODO: get userid from jwt token
	userID := int64(1)
	page := 1

	tweets, err := s.tweetStore.List(stream.Context(), userID, page)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not list tweets: %v", err)
	}
	res := &pb.ListTweetResponse{}

	for _, tweet := range tweets {
		res.Tweet = tweet
		err := stream.Send(res)
		if err != nil {
			return status.Errorf(codes.Internal, "Could not send tweet: %v", err)
		}
	}

	return nil
}

// Close store connection
func (s *Server) Close() error {
	return s.tweetStore.Close()
}
