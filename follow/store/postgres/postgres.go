package fpostgresstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/idirall22/twee/common"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
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
func (s *PostgresFollowStore) ToggleFollow(ctx context.Context, follower, followee int64) (pb.Action, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return pb.Action_UNKNOWNE_ACTION, fmt.Errorf("Could not start transaction: %v", err)
	}
	defer tx.Rollback()

	// Check if there user already follow the followee.
	var exists bool
	stmt, err := tx.PrepareContext(
		ctx, `SELECT EXISTS(SELECT 1 FROM follows WHERE followee=$1 AND follower=$2)`,
	)

	if err != nil {
		return pb.Action_UNKNOWNE_ACTION, fmt.Errorf("Could not prepare statment: %v", err)
	}

	err = stmt.QueryRowContext(ctx, followee, follower).Scan(&exists)
	if err != nil {
		return pb.Action_UNKNOWNE_ACTION, fmt.Errorf("Could not check if record already exists: %v", err)
	}

	query := "DELETE FROM follows WHERE followee=$1 AND follower=$2"
	action := pb.Action_UNKNOWNE_ACTION

	if !exists {
		query = "INSERT INTO follows (followee, follower) VALUES ($1, $2)"
		action = pb.Action_CREATED
	}

	stmt, err = tx.PrepareContext(ctx, query)
	if err != nil {
		return pb.Action_UNKNOWNE_ACTION, fmt.Errorf("Could not prepare %s follow statment: %v", action.String(), err)
	}

	_, err = stmt.ExecContext(ctx, followee, follower)
	if err != nil {
		return pb.Action_UNKNOWNE_ACTION, fmt.Errorf("Could not %s record: %v", action.String(), err)
	}

	err = tx.Commit()
	if err != nil {
		return pb.Action_UNKNOWNE_ACTION, fmt.Errorf("Could not commit transaction: %v", err)
	}

	return action, nil
}

// ListFollow followers or followee;
func (s *PostgresFollowStore) ListFollow(ctx context.Context, follower, followee int64,
	listType pb.FollowListType) ([]*pb.Follow, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not start transaction: %v", err)
	}
	defer tx.Rollback()

	// List followers
	query := "SELECT id, followee, follower FROM follows WHERE follower=$1"
	id := follower
	if listType == pb.FollowListType_FOLLOWEE {
		// List followees
		query = "SELECT id, followee, follower FROM follows WHERE followee=$1"
		id = followee
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare follow statment: %v", err)
	}

	rows, err := stmt.QueryContext(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Could not record: %v", err)
	}
	defer rows.Close()

	followList := []*pb.Follow{}
	for rows.Next() {
		f := &pb.Follow{}
		err = rows.Scan(
			&f.Id,
			&f.Followee,
			&f.Follower,
		)
		if err != nil {
			return nil, fmt.Errorf("Could not scan follow object: %v", err)
		}
		followList = append(followList, f)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("Could not commit transaction: %v", err)
	}

	return followList, nil
}
