package postgresstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/idirall22/twee/pb"
	"go.uber.org/zap"
)

// PostgresTweetStore store
type PostgresTweetStore struct {
	options *PostgresOptions
	db      *sql.DB
	logger  *zap.Logger
}

// NewPostgresTweetStore create new postgres store
func NewPostgresTweetStore(opts *PostgresOptions) (*PostgresTweetStore, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("Could not create looger: %v", err)
	}

	db, err := sql.Open("postgres", opts.String())
	if err != nil {
		attempt := 0

		for {
			time.Sleep(opts.attemptDuration)
			db, err = sql.Open("postgres", opts.String())
			if err != nil {
				logger.Info(
					fmt.Sprintf("Attempt: %d/%d --- Could not connect with database: %v",
						attempt,
						opts.GetMaxAttempt(),
						err,
					),
				)
			}
			attempt++
			if attempt >= opts.GetMaxAttempt() {
				break
			}
		}
		return nil, fmt.Errorf("Could not connect with database: %v", err)
	}

	return &PostgresTweetStore{
		options: opts,
		db:      db,
		logger:  logger,
	}, nil
}

// Create tweet
func (p *PostgresTweetStore) Create(ctx context.Context, content string) (int64, error) {
	query := fmt.Sprintf(`INSERT INTO tweets (content, user_id) VALUES ('%s', '%s')`,
		content, "1",
	)
	result, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("Could not create a record: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Could not get id of the record: %v", err)
	}

	return id, nil
}

// Update tweet
func (p *PostgresTweetStore) Update(ctx context.Context, id int64, content string) error {

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not init a transaction: %v", err)
	}

	exists := false
	stmt, err := tx.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM tweets WHERE id=$1`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prepare a statment: %v", err)
	}

	err = stmt.QueryRowContext(ctx, id).Scan(&exists)

	if err != sql.ErrNoRows && !exists {
		p.logger.Info("Could not update a tweet, Record Not exists")
		tx.Rollback()
		return fmt.Errorf("Record not exists: %v", err)
	}

	if err != nil {
		return fmt.Errorf("Could not get tweet infos: %v", err)
	}

	stmt, err = tx.PrepareContext(ctx, `UPDATE tweets SET content='%s' WHERE id=%d`)
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
func (p *PostgresTweetStore) Delete(ctx context.Context, id int64) error {

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Could not init a transaction: %v", err)
	}

	exists := false
	stmt, err := tx.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM tweets WHERE id=$1`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prepare a statment: %v", err)
	}

	err = stmt.QueryRowContext(ctx, id).Scan(&exists)

	if err != sql.ErrNoRows && !exists {
		p.logger.Info("Could not update a tweet, Record Not exists")
		tx.Rollback()
		return fmt.Errorf("Record not exists: %v", err)
	}

	if err != nil {
		return fmt.Errorf("Could not get tweet infos: %v", err)
	}

	stmt, err = tx.PrepareContext(ctx, `DELETE tweets WHERE WHERE id=%d`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Could not prepare a statment: %v", err)
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
func (p *PostgresTweetStore) Get(ctx context.Context, id int64) (*pb.Tweet, error) {

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not init a transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(
		ctx,
		"SELECT user_id, content, created_at FROM tweets WHERE id=%d",
	)

	tweet := &pb.Tweet{}
	err = stmt.QueryRowContext(ctx, id).Scan(
		&tweet.UserId,
		&tweet.Content,
		&tweet.CreatedAt,
	)

	if err != sql.ErrNoRows {
		p.logger.Info("Could not update a tweet, Record Not exists")
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
		"SELECT user_id, content, created_at FROM tweets WHERE user_id=%d LIMIT 10 OFFSET %d",
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
		err = rows.Scan(
			&tweet.UserId,
			&tweet.Content,
			&tweet.CreatedAt,
		)
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
