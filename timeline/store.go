package timeline

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store timeline interface
type Store interface {
	// List tweets
	List(ctx context.Context, userID int64, self bool, found func(tweet *pb.Tweet) error) error
}
