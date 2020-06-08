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

// New create notification
func (s *PostgresNotificationStore) New(ctx context.Context, notif *pb.Notification) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start a transaction: %v", err)
	}
	defer tx.Rollback()

	switch notif.Type {
	case pb.Notification_TWEET:
		stmt, err := tx.PrepareContext(ctx, `
			SELECT followee, follower FROM follows WHERE f.followee=$1
		`)
		if err != nil {
			return fmt.Errorf("Could not prepare select followers statment: %v", err)
		}
		rows, err := stmt.QueryContext(ctx, notif.GetTweet().GetUserId())
		if err != nil {
			return fmt.Errorf("Could not query: %v", err)
		}
		defer rows.Close()

		ids := []int64{}
		for rows.Next() {
			var followee, follower int64
			err = rows.Scan(
				&followee,
				&follower,
			)
			if err != nil {
				return fmt.Errorf("Could not scan ids: %v", err)
			}
			ids = append(ids, follower)
		}

		values := ""
		for _, id := range ids {
			values += fmt.Sprintf("('%s', '%s', '%s', %d),",
				notif.Type,
				notif.Title,
				notif.Description,
				id,
			)
		}

		values = strings.TrimRight(values, ",")

		stmt, err = tx.PrepareContext(ctx, `
			INSERT INTO notifications (type, title, description, user_id ) VALUES($1)
		`)
		if err != nil {
			return fmt.Errorf("Could not prepare statment: %v", err)
		}

		_, err = stmt.ExecContext(ctx, values)
		if err != nil {
			return fmt.Errorf("Could not create notifications: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("Could not commit transaction: %v", err)
		}
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
			SELECT data, id, user_id, type, title, description, opened
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
			&n.Data,
			&n.Id,
			&n.UserId,
			&n.Type,
			&n.Title,
			&n.Description,
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
