package tweet

import (
	"context"
	"fmt"

	eventstore "github.com/idirall22/twee/tweet/event_store"
	"github.com/idirall22/twee/tweet/store"

	"github.com/idirall22/twee/auth"

	// postgres driver
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/idirall22/twee/pb"
)

// Server server
type Server struct {
	tweetStore         store.Store
	notificationClient *pb.NotificationServiceClient
	eventStore         eventstore.EventStore
}

// NewTweetServer create new tweet server
func NewTweetServer(s store.Store, es eventstore.EventStore) (*Server, error) {
	if s == nil {
		return nil, fmt.Errorf("Store should not be NIL")
	}

	// if es == nil {
	// 	return nil, fmt.Errorf("Event Store should not be NIL")
	// }

	// go es.Start()

	return &Server{
		tweetStore: s,
		eventStore: es,
	}, nil
}

// Create a tweet.
func (s *Server) Create(ctx context.Context, req *pb.CreateTweetRequest) (*pb.CreateTweetResponse, error) {
	// get user infos from context
	userInfos, err := auth.GetUserInfosFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	content := req.Content
	if len(content) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "tweet Content is empty")
	}

	userID := userInfos.ID
	// username := userInfos.Username

	id, err := s.tweetStore.Create(ctx, userID, content)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error to create a tweet: %v", err)
	}

	res := &pb.CreateTweetResponse{
		Id: id,
	}

	// e := &pb.TweetEvent{
	// 	Action:  pb.Action_CREATED,
	// 	Title:   fmt.Sprintf("%s has just tweeted", username),
	// 	TweetId: id,
	// 	UserId:  userID,
	// }

	// go func() {
	// 	s.eventStore.Publish(ctx, e)
	// }()

	return res, nil
}

// Update a tweet.
func (s *Server) Update(ctx context.Context, req *pb.UpdateTweetRequest) (*pb.UpdateTweetResponse, error) {
	// get user infos from context
	userInfos, err := auth.GetUserInfosFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	id := req.GetId()
	content := req.GetContent()

	if id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid id")
	}

	if len(content) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "empty content")
	}

	err = s.tweetStore.Update(ctx, userInfos.ID, id, content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error to update the tweet: %v", err)
	}
	return &pb.UpdateTweetResponse{}, nil
}

// Delete a tweet.
func (s *Server) Delete(ctx context.Context, req *pb.DeleteTweetRequest) (*pb.DeleteTweetResponse, error) {
	// get user infos from context
	userInfos, err := auth.GetUserInfosFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	id := req.GetId()

	err = s.tweetStore.Delete(ctx, userInfos.ID, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not delete tweet: %v", err)
	}

	return &pb.DeleteTweetResponse{}, nil
}

// Get tweet by user id and tweet id.
func (s *Server) Get(ctx context.Context, req *pb.GetTweetRequest) (*pb.GetTweetResponse, error) {
	// get user infos from context
	userInfos, err := auth.GetUserInfosFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	id := req.GetId()
	tweet, err := s.tweetStore.Get(ctx, userInfos.ID, id)
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
	tweets, err := s.tweetStore.List(stream.Context(), req.UserId, int(req.Limit))
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
