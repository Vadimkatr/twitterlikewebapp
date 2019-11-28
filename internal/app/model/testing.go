package model

import (
	"testing"
)

func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		Username: "nickname",
		Email:    "user@gmail.com",
		Password: "password",
	}
}
