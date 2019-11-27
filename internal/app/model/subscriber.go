package model

import () // use empty import because without them I have err with import in user.go; why?


type Subscriber struct {
	Id              int `json:"id"`
	UserId          int `json:"user_id"`
	PublisherUserId int `json:"publisher_user_id"`
}
