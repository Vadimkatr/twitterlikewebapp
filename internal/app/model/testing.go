package model

import (
	"testing"
	"time"
)

// TestUser - in order not to create it many times
func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		Username: "nickname",
		Email:    "user@gmail.com",
		Password: "password",
	}
}

// TestTweet in order not to create it many times
func TestTweet(t *testing.T, u *User) *Tweet {
	t.Helper()

	return &Tweet{
		Message:  "Test tweet",
		UserId:   u.Id,
		PostTime: time.Now(),
	}
}
