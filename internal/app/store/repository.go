package store

import (
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
)

// UserRepository interface - implement methods that should be in custom repository
type UserRepository interface {
	Create(*model.User) error
	Find(int) (*model.User, error)
	FindByEmail(string) (*model.User, error)
	FindByUsername(string) (*model.User, error)
	SubscribeTo(*model.User, *model.User) error
}

// TweetRepository interface - implement methods that should be in custom repository
type TweetRepository interface {
	Create(*model.Tweet) error
	GetAllUserTweets(int) ([]string, error)
	FindTweetsFromSubscriptions(int) ([]string, error)
}
