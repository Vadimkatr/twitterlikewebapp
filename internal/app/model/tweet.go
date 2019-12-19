package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Tweet struct {
	Id       int       `json:"id"`
	Message  string    `json:"message"`
	PostTime time.Time `json:"post_time"`
	UserId   int       `json:"user_id"`
}

// Validate Tweet fields
func (t *Tweet) Validate() error {
	return validation.ValidateStruct(
		t,
		validation.Field(&t.Message, validation.Required, validation.Length(1, 280)),
	)
}
