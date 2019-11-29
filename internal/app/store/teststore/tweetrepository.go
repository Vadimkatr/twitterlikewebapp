package teststore

import (
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
)

type TweetRepository struct {
	store  *Store
	tweets map[int]*model.Tweet
}

func (r *TweetRepository) Create(t *model.Tweet) error {

	if err := t.Validate(); err != nil {
		return err
	}

	t.Id = len(r.tweets) + 1
	r.tweets[t.Id] = t

	return nil
}

func (r *TweetRepository) FindTweetsFromSubscriptions(id int) ([]string, error) {
	// TODO: init method
	return []string{""}, nil
}

func (r *TweetRepository) GetAllUserTweets(userId int) ([]string, error) {
	u, err := r.store.User().Find(userId)
	if err != nil {
		return []string{}, store.ErrRecordNotFound
	}
	var tweets []string
	for _, t := range r.tweets {
		if t.UserId == u.Id {
			tweets = append(tweets, t.Message)
		}
	}

	return tweets, nil
}
