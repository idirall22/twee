package postgresstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/idirall22/twee/common"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
	"go.uber.org/zap"
)

// PostgresTweetStore store
type PostgresTweetStore struct {
	options *option.PostgresOptions
	db      *sql.DB
	logger  *zap.Logger
}

// NewPostgresTweetStore create new postgres store
func NewPostgresTweetStore(opts *option.PostgresOptions) (*PostgresTweetStore, error) {
	logger, db, err := common.SetupPostgres(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to db: %v", err)
	}

	return &PostgresTweetStore{
		options: opts,
		db:      db,
		logger:  logger,
	}, nil
}

// Create tweet
func (p *PostgresTweetStore) Create(ctx context.Context, userID int64, content string) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO tweets (content, user_id)
		VALUES ('%s', '%s')
		RETURNING  id`,
		content, "1",
	)
	var id int64
	err := p.db.QueryRowContext(ctx, query).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("Could not create a record: %v", err)
	}
	// id, err := result.RowsAffected()
	// if err != nil {
	// 	return 0, fmt.Errorf("Could not get id of the record: %v", err)
	// }

	return id, nil
}

// Update tweet
func (p *PostgresTweetStore) Update(ctx context.Context, userID int64, id int64, content string) error {

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not init a transaction: %v", err)
	}

	var exists bool
	stmt, err := tx.PrepareContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM tweets WHERE id=$1)",
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prepare a statment: %v", err)
	}

	err = stmt.QueryRowContext(ctx, id).Scan(&exists)

	if err != sql.ErrNoRows && !exists {
		tx.Rollback()
		return fmt.Errorf("Record not exists: %v", err)
	}

	if err != nil {
		return fmt.Errorf("Could not get tweet infos: %v", err)
	}

	stmt, err = tx.PrepareContext(ctx, `UPDATE tweets SET content=$1 WHERE id=$2`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prepare a statment: %v", err)
	}

	_, err = stmt.ExecContext(ctx, content, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not create a record: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not commit transaction: %v", err)
	}
	return nil
}

// Delete tweet
func (p *PostgresTweetStore) Delete(ctx context.Context, userID int64, id int64) error {

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not init a transaction: %v", err)
	}

	var exists bool
	stmt, err := tx.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM tweets WHERE id=$1)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prepare a statment: %v", err)
	}

	err = stmt.QueryRowContext(ctx, id).Scan(&exists)

	if !exists {
		return fmt.Errorf("Record not exists")
	}

	if err == sql.ErrNoRows {
		p.logger.Info("Could not delete a tweet, Record Not exists")
		tx.Rollback()
		return fmt.Errorf("Record not exists: %v", err)
	}

	if err != nil {
		return fmt.Errorf("Could not get tweet infos: %v", err)
	}

	stmt, err = tx.PrepareContext(ctx, `DELETE FROM tweets WHERE id=$1`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prepare delete statment: %v", err)
	}

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not delete a record: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not commit transaction: %v", err)
	}

	return nil
}

// Get tweet
func (p *PostgresTweetStore) Get(ctx context.Context, userID int64, id int64) (*pb.Tweet, error) {

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not init a transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(
		ctx,
		"SELECT user_id, content, created_at FROM tweets WHERE id=$1",
	)

	tweet := &pb.Tweet{}
	var t time.Time

	err = stmt.QueryRowContext(ctx, id).Scan(
		&tweet.UserId,
		&tweet.Content,
		&t,
	)
	tweet.CreatedAt, _ = ptypes.TimestampProto(t)

	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, fmt.Errorf("Record not exists: %v", err)
	}

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Could not get tweet: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Could not commit transaction: %v", err)
	}

	return tweet, nil
}

// List tweets
func (p *PostgresTweetStore) List(ctx context.Context, userID int64, page int) ([]*pb.Tweet, error) {

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not init a transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(
		ctx,
		"SELECT id, user_id, content, created_at FROM tweets WHERE user_id=$1 LIMIT 10 OFFSET $2",
	)

	tweets := []*pb.Tweet{}
	rows, err := stmt.QueryContext(ctx, userID, page)

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Could not get tweets")
	}
	defer rows.Close()

	for rows.Next() {
		tweet := &pb.Tweet{}
		var t time.Time

		err = rows.Scan(
			&tweet.Id,
			&tweet.UserId,
			&tweet.Content,
			&t,
		)
		tweet.CreatedAt, _ = ptypes.TimestampProto(t)

		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("Could not scan: %v", err)
		}
		tweets = append(tweets, tweet)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Could not commit transaction: %v", err)
	}

	return tweets, nil
}

// Close close connections
func (p *PostgresTweetStore) Close() error {
	p.logger.Sync()
	p.db.Close()
	return nil
}
