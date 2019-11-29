package model

import (
	"testing"
	"time"
)

func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		Username: "nickname",
		Email:    "user@gmail.com",
		Password: "password",
	}
}

func TestTweet(t *testing.T) *Tweet {
	t.Helper()

	return &Tweet{
		Message: "Test tweet",
		UserId:   0,
		PostTime: time.Now(),
	}
}
