package teststore_test

import (
	"testing"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store/teststore"
	"github.com/stretchr/testify/assert"
)

func TestTweetRepository_Create(t *testing.T) {
	s := teststore.New()
	tw := model.TestTweet(t, model.TestUser(t))
	assert.NoError(t, s.Tweet().Create(tw))
	assert.NotNil(t, tw.Id)
}

func TestTweetRepository_GetAllUserTweets(t *testing.T) {
	// try to find tweet, but now tweet.UserId in store => err
	s := teststore.New()
	tw := model.TestTweet(t, model.TestUser(t))
	_, err := s.Tweet().GetAllUserTweets(tw.UserId)
	assert.EqualError(t, err, store.ErrUserNotFound.Error())

	// add user to store; set tweet.UserID as user.Id and create tweet; find user tweet
	u1 := model.TestUser(t)
	s.User().Create(u1)
	u2, err := s.User().FindByEmail(u1.Email)
	tw1 := model.TestTweet(t, u2)
	s.Tweet().Create(tw1)
	tweets, err := s.Tweet().GetAllUserTweets(tw1.UserId)

	assert.NoError(t, err)
	assert.Equal(t, []string{tw1.Message}, tweets)
}
