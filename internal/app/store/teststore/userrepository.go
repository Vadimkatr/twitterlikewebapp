package teststore

import (
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
)

// UserRepository ...
type UserRepository struct {
	store *Store
	users map[int]*model.User
}

// Create ...
func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	u.Id = len(r.users) + 1
	r.users[u.Id] = u

	return nil
}

// Find ...
func (r *UserRepository) Find(id int) (*model.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, store.ErrUserNotFound
	}

	return u, nil
}

// FindByEmail ...
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, store.ErrUserNotFound
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}

	return nil, store.ErrUserNotFound
}

func (r *UserRepository) SubscribeTo(u *model.User, su *model.User) error {
	// all magic runs on server, so we cant test this func there
	return nil
}
