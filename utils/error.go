package utils

import "fmt"

var (
	// ErrNotExists record not exists
	ErrNotExists = fmt.Errorf("Record not exists")

	// ErrInvalidID id is not valid
	ErrInvalidID = fmt.Errorf("Invalid id")

	// ErrInvalidUserID user id is not valid
	ErrInvalidUserID = fmt.Errorf("Invalid user id")

	// ErrUpdate could not update a record
	ErrUpdate = fmt.Errorf("Could not update a record")

	// ErrCreate could not create a record
	ErrCreate = fmt.Errorf("Could not create a record")

	// ErrUUID error to generate uuid
	ErrUUID = fmt.Errorf("Could not create uuid")

	// ErrUserRecordNotExists when user record not exists
	ErrUserRecordNotExists = fmt.Errorf("user record not exists")
)
