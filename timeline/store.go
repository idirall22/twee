package timeline

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store timeline interface
type Store interface {
	// List tweets from user timeline
	List(ctx context.Context, userID int64, found func(tweet *pb.Tweet) error) error
}
