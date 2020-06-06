package ustore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/idirall22/twee/common"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
)

// PostgresUserStore struct
type PostgresUserStore struct {
	options *option.PostgresOptions
	db      *sql.DB
}

// NewPostgresUserStore create new postgres store.
func NewPostgresUserStore(opts *option.PostgresOptions) (*PostgresUserStore, error) {
	_, db, err := common.SetupPostgres(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to db: %v", err)
	}

	return &PostgresUserStore{
		options: opts,
		db:      db,
	}, nil
}

// List users profile
func (s *PostgresUserStore) List(ctx context.Context, limit, offset int32, found func(user *pb.User) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
	SELECT id, username, followee_count, follower_count
	FROM users LIMIT $1 OFFSET $2
	`)

	if err != nil {
		return fmt.Errorf("Could not prepare SELECT statment: %v", err)
	}

	rows, err := stmt.QueryContext(ctx, limit, offset)
	if err != nil {
		return fmt.Errorf("Could not fetch users: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &pb.User{}
		err := rows.Scan(
			&user.Id,
			&user.Username,
			&user.FolloweeCount,
			&user.FollowerCount,
		)

		if err != nil {
			return fmt.Errorf("Could not scan data: %v", err)
		}

		err = found(user)
		if err != nil {
			return fmt.Errorf("Could not stream data: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Could not commit transaction: %v", err)
	}
	return nil
}

// Profile Get user profile by username
func (s *PostgresUserStore) Profile(ctx context.Context, username string) (*pb.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not start transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		SELECT id, username, followee_count, follower_count
		FROM users LIMIT 1
	`)

	if err != nil {
		return nil, fmt.Errorf("Could not prepare SELECT statment: %v", err)
	}

	user := &pb.User{}
	err = stmt.QueryRowContext(ctx).Scan(
		&user.Id,
		&user.Username,
		&user.FolloweeCount,
		&user.FollowerCount,
	)

	if err != nil {
		return nil, fmt.Errorf("Could not query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("Could not commit transaction: %v", err)
	}
	return user, nil
}
