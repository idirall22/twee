package store

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store store notification interface.
type Store interface {
	// New create notification
	NewTweetNotification(ctx context.Context, followersList []*pb.Follow, notif *pb.TweetEvent,
		cNotification chan<- *pb.Notification) error
	// List notifications
	List(ctx context.Context, userID int64, found func(n *pb.Notification) error) error
}
