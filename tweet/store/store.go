package store

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store interface
type Store interface {
	// create tweet
	Create(ctx context.Context, userID int64, content string) (int64, error)
	// update tweet
	Update(ctx context.Context, userID int64, id int64, content string) error
	// delete tweet
	Delete(ctx context.Context, userID int64, id int64) error
	// get tweet
	Get(ctx context.Context, userID int64, id int64) (*pb.Tweet, error)
	// list tweets
	List(ctx context.Context, userID int64, page int) ([]*pb.Tweet, error)
	// Close
	Close() error
}
