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
func (s *PostgresNotificationStore) New(
	ctx context.Context,
	nn *pb.NewNotification,
	notifChan chan<- *pb.Notification,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not start a transaction: %v", err)
	}
	defer tx.Rollback()

	switch nn.Type {
	case pb.Type_TWEET:
		stmt, err := tx.PrepareContext(ctx, `
			SELECT followee, follower FROM follows WHERE followee=$1
		`)
		if err != nil {
			return fmt.Errorf("Could not prepare select followers statment: %v", err)
		}
		rows, err := stmt.QueryContext(ctx, nn.UserOrigin)
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

		if len(ids) == 0 {
			return nil
		}

		values := ""
		for _, id := range ids {
			values += fmt.Sprintf("(%d, '%s', %d, '%s', %d, %v),",
				nn.UserOrigin,
				nn.Type,
				nn.TypeId,
				nn.Title,
				id,
				nn.Opened,
			)
		}
		values = strings.TrimRight(values, ",")

		query := fmt.Sprintf(`
			INSERT INTO notifications
			(user_origin, type, type_id, title, user_id, opened) VALUES %s RETURNING id, user_id`,
			values,
		)

		stmt, err = tx.PrepareContext(ctx, query)

		if err != nil {
			return fmt.Errorf("Could not prepare insert statment: %v", err)
		}

		rows, err = stmt.QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("Could not create notifications: %v", err)
		}
		defer rows.Close()

		notif := &pb.Notification{

			UserOrigin: nn.UserOrigin,
			Type:       nn.Type,
			TypeId:     nn.TypeId,
			Title:      nn.Title,
			Opened:     nn.Opened,
		}

		for rows.Next() {
			err = rows.Scan(
				&notif.Id,
				&notif.UserId,
			)
			if err != nil {
				return fmt.Errorf("Could not scan id: %v", err)
			}
			notifChan <- notif
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
