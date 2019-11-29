package teststore

import (
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
)

// Store ...
type Store struct {
	userRepository  *UserRepository
	tweetRepository *TweetRepository
}

// New ...
func New() *Store {
	return &Store{}
}

// User ...
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
		users: make(map[int]*model.User),
	}

	return s.userRepository
}

func (s *Store) Tweet() store.TweetRepository {
	if s.tweetRepository != nil {
		return s.tweetRepository
	}

	s.tweetRepository = &TweetRepository{
		store:  s,
		tweets: make(map[int]*model.Tweet),
	}

	return s.tweetRepository
}
