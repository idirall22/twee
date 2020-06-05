package tlstore

import (
	"context"
	"database/sql"
	"fmt"

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
	found func(tm *pb.Timeline) error,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(
		ctx,
		`
			SELECT tm.tweet_id
			FROM timeline as tm
			LEFT JOIN (SELECT content, user_id FROM tweets WHERE id=tm.tweet_id)
			ON tm.follower=$1
		`,
	)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prempare statment: %v", err)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not Query timeline tweets: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		tm := &pb.Timeline{}
		tweet := &pb.Tweet{}

		rows.Scan(
			&tweet.Id,
			&tweet.Content,
			&tm.Followee
		)
	}

	return nil
}
