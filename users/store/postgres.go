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
func (s *PostgresUserStore) List(ctx context.Context, limit, offset int32, found func(profile *pb.Profile) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start transaction: %v", err)
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `
		SELECT
		p.id, p.user_id, p.followee_count, p.follower_count, u.id, u.username
		FROM profiles AS p
		INNER JOIN users AS u ON p.user_id=u.id
		LIMIT $1 OFFSET $2
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
		profile := &pb.Profile{}
		user := &pb.User{}
		err := rows.Scan(
			&profile.Id,
			&profile.UserId,
			&profile.FolloweeCount,
			&profile.FollowerCount,
			&user.Id,
			&user.Username,
		)
		profile.User = user

		if err != nil {
			return fmt.Errorf("Could not scan data: %v", err)
		}

		err = found(profile)
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
func (s *PostgresUserStore) Profile(ctx context.Context, username string) (*pb.Profile, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not start transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		SELECT p.id, p.user_id, p.followee_count, p.follower_count, u.id, u.username
		FROM (SELECT * FROM users WHERE username=$1 LIMIT 1) as u
		INNER JOIN (SELECT * FROM profiles WHERE u.id=user_id) as p
	`)

	if err != nil {
		return nil, fmt.Errorf("Could not prepare SELECT statment: %v", err)
	}

	profile := &pb.Profile{}
	user := &pb.User{}
	err = stmt.QueryRowContext(ctx).Scan(
		&profile.Id,
		&profile.UserId,
		&profile.FolloweeCount,
		&profile.FollowerCount,
		&user.Id,
		&user.Username,
	)
	profile.User = user

	if err != nil {
		return nil, fmt.Errorf("Could not query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("Could not commit transaction: %v", err)
	}
	return profile, nil
}
