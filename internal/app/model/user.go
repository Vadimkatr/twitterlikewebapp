package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

)

type User struct {
	Account_id        int    `json:"id"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

// Validate User fields
func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Password, validation.Required, validation.Length(6, 100)),
	)
}
