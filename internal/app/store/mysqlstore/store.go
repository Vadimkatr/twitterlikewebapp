package mysqlstore

import (
	"database/sql"
	
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
)

type Store struct {
	db              *sql.DB
	userRepository  *UserRepository
	tweetRepository *TweetRepository

}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}

func (s *Store) Tweet() store.TweetRepository {
	if s.tweetRepository != nil {
		return s.tweetRepository
	}

	s.tweetRepository = &TweetRepository{
		store: s,
	}

	return s.tweetRepository
}
