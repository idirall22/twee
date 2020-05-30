package auth

import (
	"context"

	"github.com/idirall22/twee/pb"
)

// Store auth store interface
type Store interface {
	// Create new user
	Create(ctx context.Context, username, hashPassword string) error
	// Find user by username
	Find(ctx context.Context, username string) (*pb.User, error)
	// List users
	List(ctx context.Context, page int) ([]*pb.User, error)
	// Close connection
	Close() error
}
