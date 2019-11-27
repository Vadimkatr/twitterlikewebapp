package mysqlstore

import (
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	return r.store.db.QueryRow(
		"INSERT INTO users (email, username, encrypted_password) VALUES (?, ?) RETURNING id",
		u.Email,
		u.Username,
		u.EncryptedPassword,
	).Scan(&u.Account_id)
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	return u, nil
}
