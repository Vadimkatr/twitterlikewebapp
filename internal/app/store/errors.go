package store

import (
	"errors"
)

var (
	// ErrRecordNotFound something
	ErrRecordNotFound = errors.New("record not found")
	ErrRecordIsExist = errors.New("record is exist")
)
