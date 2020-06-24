package store

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// TimelineStore timeline interface
type TimelineStore interface {
	// List tweets
	List(ctx context.Context, userID int64, followList []*pb.Follow, self pb.TimelineType, found func(tweet *pb.Tweet) error) error
}
