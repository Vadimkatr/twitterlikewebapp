package store

// Store interface - implement methods that should be in custom store
type Store interface {
	// use to get access to repository methods in code from store
	User() UserRepository
	Tweet() TweetRepository
}
