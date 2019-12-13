package store

import (
	"errors"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserIsExist         = errors.New("user is exist")

	ErrSubscritionIsCreate = errors.New("subscription is already create")
)
