package store

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// FollowStore interface
type FollowStore interface {
	// ToggleFollow follow, unfollow a user
	ToggleFollow(ctx context.Context, follower, followee int64) (pb.Action, error)

	// List followers or followee;
	ListFollow(ctx context.Context, follower, followee int64, listType pb.FollowListType) ([]*pb.Follow, error)
}
