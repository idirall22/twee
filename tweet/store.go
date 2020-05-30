package tweet

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store interface
type Store interface {
	// create tweet
	Create(ctx context.Context, content string) (int64, error)
	// update tweet
	Update(ctx context.Context, tweet *pb.Tweet) error
	// delete tweet
	Delete(ctx context.Context, id int64) error
	// get tweet
	Get(ctx context.Context, id int64) (*pb.Tweet, error)
	// list tweets
	List(ctx context.Context, page int) ([]*pb.Tweet, error)
	// Close
	Close() error
}
