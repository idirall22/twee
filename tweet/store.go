package tweet

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store interface
type Store interface {
	// create tweet
	Create(ctx context.Context, tweet *pb.Tweet) (string, error)
	// update tweet
	Update(ctx context.Context, tweet *pb.Tweet) error
	// delete tweet
	Delete(ctx context.Context, id, userID string) error
	// get tweet
	Get(ctx context.Context, id, userID string) (*pb.Tweet, error)
	// list tweets
	List(ctx context.Context, filter *pb.TweetFilter, found func() error) error
}
