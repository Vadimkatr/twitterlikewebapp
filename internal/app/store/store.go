package store

type Store interface {
	User()  UserRepository
	Tweet() TweetRepository
	Subscriber() SubscriberRepository
}
