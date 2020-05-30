package postgresstore

import (
	"fmt"
	"time"
)

// PostgresOptions struct
type PostgresOptions struct {
	host            string
	port            int
	user            string
	password        string
	dbname          string
	maxAttempt      int
	attemptDuration time.Duration
}

// NewPostgresOptions create postgres options
func NewPostgresOptions(host, user, password, dbname string,
	maxAttempt, port int, attemptDuration time.Duration) *PostgresOptions {
	return &PostgresOptions{
		host:            host,
		port:            port,
		user:            user,
		password:        password,
		dbname:          dbname,
		maxAttempt:      maxAttempt,
		attemptDuration: attemptDuration,
	}
}

// String return connection string
func (o *PostgresOptions) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		o.host, o.port, o.user, o.password, o.dbname)
}

// GetMaxAttempt get max attempt
func (o *PostgresOptions) GetMaxAttempt() int {
	return o.maxAttempt
}

// GetAttemptDuration get attempt duration
func (o *PostgresOptions) GetAttemptDuration() time.Duration {
	return o.attemptDuration
}
