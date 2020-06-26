package nstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/idirall22/twee/common"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
)

// PostgresNotificationStore postgres notification store
type PostgresNotificationStore struct {
	options *option.PostgresOptions
	db      *sql.DB
}

// NewPostgresNotificationStore create new postgres notification store
func NewPostgresNotificationStore(opts *option.PostgresOptions) (*PostgresNotificationStore, error) {
	_, db, err := common.SetupPostgres(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to db: %v", err)
	}

	return &PostgresNotificationStore{
		options: opts,
		db:      db,
	}, nil
}

// NewTweetNotification create tweet notification
func (s *PostgresNotificationStore) NewTweetNotification(
	ctx context.Context,
	followersList []*pb.Follow,
	te *pb.TweetEvent,
	notifChan chan<- *pb.Notification,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start a transaction: %v", err)
	}
	defer tx.Rollback()

	args := []string{}
	for _, follow := range followersList {
		args = append(args,
			fmt.Sprintf("(%d, '%s', %d, '%s', %d, false)",
				te.UserId, pb.Type_TWEET, te.TweetId, te.Title, follow.Follower),
		)
	}
	query := fmt.Sprintf(
		"INSERT INTO notifications (user_origin, type, type_id, title, user_id, opened) values %s",
		strings.Join(args, ","),
	)
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Could not commit transaction: %v", err)
	}
	return nil
}

// List notifications
func (s *PostgresNotificationStore) List(
	ctx context.Context,
	userID int64,
	found func(n *pb.Notification) error,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start a transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(
		ctx,
		`
			SELECT id, user_origin, type, type_id, title, user_id, opened
			FROM notifications WHERE user_id=$1
		`,
	)
	if err != nil {
		return fmt.Errorf("Could not prepare statment: %v", err)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return fmt.Errorf("Could not prepare statment: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		n := &pb.Notification{}
		err = rows.Scan(
			&n.Id,
			&n.UserOrigin,
			&n.Type,
			&n.TypeId,
			&n.Title,
			&n.UserId,
			&n.Opened,
		)
		if err != nil {
			return fmt.Errorf("Could not scan notification: %v", err)
		}

		err = found(n)
		if err != nil {
			return fmt.Errorf("Could not send notification: %v", err)
		}
	}
	return nil
}
