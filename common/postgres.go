package common

import (
	"database/sql"
	"fmt"
	"time"

	option "github.com/idirall22/twee/options"

	"go.uber.org/zap"
)

// SetupPostgres postgres connection
func SetupPostgres(opts *option.PostgresOptions) (*zap.Logger, *sql.DB, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not create looger: %v", err)
	}

	db, err := sql.Open("postgres", opts.String())
	if err != nil {
		attempt := 0

		for {
			time.Sleep(opts.GetAttemptDuration())
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
		return nil, nil, fmt.Errorf("Could not connect with database: %v", err)
	}

	return logger, db, nil
}
