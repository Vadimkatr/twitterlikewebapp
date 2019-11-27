package store

import (
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
)

type UserRepository interface {
	Create(*model.User) error
	FindByEmail(string) (*model.User, error)
	Find(int) (*model.User, error)
}

type TweetRepository interface {
	Create(*model.Tweet) error
}