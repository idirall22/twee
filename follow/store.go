package follow

import "context"

// Store interface
type Store interface {
	// Follow follow a user
	Follow(ctx context.Context, followee int64)
	// Unfollow unfollow a user
	Unfollow(ctx context.Context, followee int64)
}
