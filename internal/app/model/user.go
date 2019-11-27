package model

type User struct {
	Account_id        int    `json:"id"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}
