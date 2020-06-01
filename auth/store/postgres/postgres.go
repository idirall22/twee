package apstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/idirall22/twee/common"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
	"go.uber.org/zap"
)

// PostgresAuthStore auth postgres store struct
type PostgresAuthStore struct {
	options *option.PostgresOptions
	db      *sql.DB
	logger  *zap.Logger
}

// NewPostgresAuthStore create new auth postgres store
func NewPostgresAuthStore(opts *option.PostgresOptions) (*PostgresAuthStore, error) {

	logger, db, err := common.SetupPostgres(opts)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to db: %v", err)
	}

	return &PostgresAuthStore{
		options: opts,
		db:      db,
		logger:  logger,
	}, nil
}

// Create new user
func (s *PostgresAuthStore) Create(ctx context.Context, username, hashPassword string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error to start transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE username=$1);",
	)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error to prepare stmt: %v", err)
	}

	var exists bool
	err = stmt.QueryRowContext(ctx, username).Scan(&exists)

	if exists || err == sql.ErrNoRows {
		tx.Rollback()
		return fmt.Errorf("User Already exists")
	}

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error to check if record already exists: %v", err)
	}

	stmt, err = tx.PrepareContext(
		ctx,
		"INSERT INTO users (username, hash_password) values ($1, $2)",
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error to prepare stmt: %v", err)
	}

	_, err = stmt.ExecContext(ctx, username, hashPassword)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error to execute query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error to commit transaction: %v", err)
	}

	return nil
}

// Find user by username
func (s *PostgresAuthStore) Find(ctx context.Context, username string) (*pb.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Error to start transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(ctx, "SELECT id, hash_password FROM users WHERE username=$1")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Error to prepare stmt: %v", err)
	}

	user := &pb.User{
		Username: username,
	}
	err = stmt.QueryRowContext(ctx, username).Scan(&user.Id, &user.HashPassword)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Error to execute query: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("Error to commit transaction: %v", err)
	}

	return user, nil
}

// List users
func (s *PostgresAuthStore) List(ctx context.Context, page int) ([]*pb.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Error to start transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(ctx, "SELECT id, username FROM users LIMIT 10 OFFSET $1")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Error to prepare stmt: %v", err)
	}

	rows, err := stmt.QueryContext(ctx, page)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Error to query users: %v", err)
	}
	defer rows.Close()

	users := []*pb.User{}
	for rows.Next() {
		user := &pb.User{}
		err = rows.Scan(
			user.Id,
			user.Username,
		)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("Error to scan user: %v", err)
		}
		users = append(users, user)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("Error to commit transaction: %v", err)
	}

	return users, nil
}

// Close close connection
func (s *PostgresAuthStore) Close() error {
	return s.db.Close()
}
