package common

import (
	"time"

	option "github.com/idirall22/twee/options"
)

var (
	// PostgresTestOptions postgres database options for testing.
	PostgresTestOptions = option.NewPostgresOptions(
		"0.0.0.0",
		"postgres",
		"password",
		"twee",
		3,
		5432,
		time.Second,
	)
)
