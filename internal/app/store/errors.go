package store

import (
	"errors"
)

var (
	// ErrUserNotFound something
	ErrUserNotFound = errors.New("user not found")
	ErrUserIsExist = errors.New("user is exist")
)
