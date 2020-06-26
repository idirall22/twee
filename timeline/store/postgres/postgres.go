package tlpostgresstore

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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
	followList []*pb.Follow,
	timelineType pb.TimelineType,
	found func(tm *pb.Tweet) error,
) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start transaction: %v", err)
	}

	defer tx.Rollback()

	usersString := strconv.FormatInt(userID, 10)
	query := "SELECT * FROM tweets WHERE user_id = ANY($1::int[])"
	usersString = common.GetFolloweeString(usersString, followList)
	// if timelineType == pb.TimelineType_HOME {
	// }

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("Could not prepare statment: %v", err)
	}

	rows, err := stmt.QueryContext(ctx, usersString)
	if err != nil {
		return fmt.Errorf("Could not Query timeline tweets: %v", err)
	}

	defer rows.Close()

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
	return nil
}

func getFolloweeString(in string, followList []*pb.Follow) string {
	out := in + ","
	for _, f := range followList {
		out += strconv.FormatInt(f.Followee, 10) + ","
	}
	out = strings.TrimRight(out, ",")
	return out
}
