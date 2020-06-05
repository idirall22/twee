package fstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/idirall22/twee/common"
	option "github.com/idirall22/twee/options"
)

// PostgresFollowStore follow postgres store struct.
type PostgresFollowStore struct {
	options *option.PostgresOptions
	db      *sql.DB
}

// NewPostgresFollowStore create new follow postgres store
func NewPostgresFollowStore(opts *option.PostgresOptions) (*PostgresFollowStore, error) {
	_, db, err := common.SetupPostgres(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to db: %v", err)
	}

	return &PostgresFollowStore{
		options: opts,
		db:      db,
	}, nil
}

// ToggleFollow toggle folow a user.
func (s *PostgresFollowStore) ToggleFollow(ctx context.Context, follower, followee int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start transaction: %v", err)
	}
	defer tx.Rollback()

	var exists bool

	stmt, err := tx.PrepareContext(
		ctx, `SELECT EXISTS(SELECT 1 FROM follows WHERE followee=$1 AND follower=$2)`,
	)

	if err != nil {
		return fmt.Errorf("Could not prepare statment: %v", err)
	}

	err = stmt.QueryRowContext(ctx, followee, follower).Scan(&exists)
	if err != nil {
		return fmt.Errorf("Could not check if record already exists: %v", err)
	}

	if exists {
		stmt, err = tx.PrepareContext(
			ctx, `DELETE FROM follows WHERE followee=$1 AND follower=$2`,
		)
		if err != nil {
			return fmt.Errorf("Could not prepare delete follow statment: %v", err)
		}
	} else {
		stmt, err = tx.PrepareContext(
			ctx, `INSERT INTO follows (followee, follower) VALUES ($1, $2)`,
		)
		if err != nil {
			return fmt.Errorf("Could not prepare insert follow statment: %v", err)
		}
	}

	_, err = stmt.ExecContext(ctx, followee, follower)
	if err != nil {
		return fmt.Errorf("Could not insert/delete record: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Could not commit transaction: %v", err)
	}

	return nil
}
