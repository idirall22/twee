package user

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store user interface
type Store interface {
	// List users profile
	List(ctx context.Context, limit, offset int32, found func(profile *pb.Profile) error) error
	// Get user profile by username
	Profile(ctx context.Context, username string) (*pb.Profile, error)
}
