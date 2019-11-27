package mysqlstore

import (
	"database/sql"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
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

	row, err := r.store.db.Query(
		"INSERT INTO users (email, username, encrypted_password) VALUES (?, ?, ?)",
		u.Email,
		u.Username,
		u.EncryptedPassword,
	)

	defer row.Close()
	if err != nil {
		return err
	}
	for row.Next() {
		err := row.Scan(&u.Id)
		return err
	}
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, email, username, encrypted_password FROM users WHERE email = ?", 
		email,
	).Scan(
		&u.Id, 
		&u.Email, 
		&u.Username,
		&u.EncryptedPassword,
	); err != nil {
		if err != sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	
	return u, nil
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, email, username, encrypted_password FROM users WHERE id = ?", 
		id,
	).Scan(
		&u.Id, 
		&u.Email, 
		&u.Username,
		&u.EncryptedPassword,
	); err != nil {
		if err != sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	
	return u, nil
}
