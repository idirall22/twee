package tlstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/idirall22/twee/common"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
)

// PostgresTimelineStore timeline postgres store struct
type PostgresTimelineStore struct {
	options *option.PostgresOptions
	db      *sql.DB
}

// NewPostgresTimelineStore create new Timeline postgres store
func NewPostgresTimelineStore(opts *option.PostgresOptions) (*PostgresTimelineStore, error) {
	_, db, err := common.SetupPostgres(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to db: %v", err)
	}

	return &PostgresTimelineStore{
		options: opts,
		db:      db,
	}, nil
}

// List timeline user tweets
func (s *PostgresTimelineStore) List(
	ctx context.Context,
	userID int64,
	self bool,
	found func(tm *pb.Tweet) error,
) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start transaction: %v", err)
	}

	defer tx.Rollback()

	var stmt *sql.Stmt

	if self {
		stmt, err = tx.PrepareContext(
			ctx,
			`
				SELECT * FROM tweets WHERE user_id=$1 LIMIT $2 OFFSET $3
			`,
		)
		if err != nil {
			return fmt.Errorf("Could not prempare statment: %v", err)
		}
	} else {
		stmt, err = tx.PrepareContext(
			ctx,
			`
				SELECT f.followee, f.follower, t.id, t.user_id, t.content
				FROM follows AS f INNER JOIN tweets AS t
				ON f.follower=$1
			`,
		)
		if err != nil {
			return fmt.Errorf("Could not prempare statment: %v", err)
		}
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return fmt.Errorf("Could not Query timeline tweets: %v", err)
	}

	defer rows.Close()

	if self {
		for rows.Next() {
			tweet := &pb.Tweet{}
			var t time.Time
			rows.Scan(
				&tweet.Id,
				&tweet.UserId,
				&tweet.Content,
				&t,
			)
			tweet.CreatedAt, _ = ptypes.TimestampProto(t)
			err = found(tweet)
			if err != nil {
				return fmt.Errorf("Could not send tweet: %v", err)
			}
		}
	} else {
		var followee int
		var follower int
		for rows.Next() {
			tweet := &pb.Tweet{}
			var t time.Time
			rows.Scan(
				&followee,
				&follower,
				&tweet.Id,
				&tweet.UserId,
				&tweet.Content,
				&t,
			)
			tweet.CreatedAt, _ = ptypes.TimestampProto(t)
			err = found(tweet)
			if err != nil {
				return fmt.Errorf("Could not send tweet: %v", err)
			}
		}
	}

	return nil
}
