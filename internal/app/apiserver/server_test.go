package apiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store/teststore"
)

func TestServer_HandleUsersCreate(t *testing.T) {
	s := newServer(teststore.New())
	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]interface{}{
				"email":    "user@example.org",
				"password": "some_password",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "invalid payload",
			payload:      "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid params",
			payload: map[string]interface{}{
				"email":    "invalid",
				"password": "short",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/register", b)
			s.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_HandleUsersLogin(t *testing.T) {
	s := newServer(teststore.New())

	u := model.TestUser(t)
	s.store.User().Create(u)

	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]interface{}{
				"email":    u.Email,
				"password": u.Password,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid payload",
			payload:      "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid params",
			payload: map[string]interface{}{
				"email":    "invalid",
				"password": "short",
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/login", b)
			s.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_HandleTweetsCreate(t *testing.T) {
	s := newServer(teststore.New())

	u := model.TestUser(t)
	tw := model.TestTweet(t, u)

	s.store.User().Create(u)

	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]interface{}{
				"message": tw.Message,
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "invalid payload",
			payload:      "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid params",
			payload: map[string]interface{}{
				"some_param": "invalid",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/tweets", b)
			req.Header.Set("user_id", strconv.Itoa(u.Id))
			s.handleTweetsCreate().ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_HandleGetAllTweetsFromSubscriptions(t *testing.T) {
	s := newServer(teststore.New())

	u := model.TestUser(t)
	u2 := model.TestUser(t)
	u2.Username = "test_user"
	u2.Email = "test@gmail.com"

	tw2 := model.TestTweet(t, u2)

	s.store.User().Create(u)
	s.store.User().Create(u2)
	s.store.Tweet().Create(tw2)
	s.store.User().SubscribeTo(u, u2)
	
	testCases := []struct {
		name         string
		idFromHeader string
		response     map[string]string
		expectedCode int
	}{
		{
			name:         "valid",
			idFromHeader: strconv.Itoa(u.Id),
			response:     map[string]string{"tweets": tw2.Message},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid",
			idFromHeader: "-1",
			response:     nil,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/tweets", nil)
			req.Header.Set("user_id", tc.idFromHeader)
			s.handleGetAllTweetsFromSubscriptions().ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_HandleGetAllUserTweets(t *testing.T) {
	s := newServer(teststore.New())

	u := model.TestUser(t)
	tw := model.TestTweet(t, u)

	s.store.User().Create(u)
	s.store.Tweet().Create(tw)
	
	testCases := []struct {
		name         string
		idFromHeader string
		response     map[string]string
		expectedCode int
	}{
		{
			name:         "valid",
			idFromHeader: strconv.Itoa(u.Id),
			response:     map[string]string{"tweets": tw.Message},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid",
			idFromHeader: "-1",
			response:     nil,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/mytweets", nil)
			req.Header.Set("user_id", tc.idFromHeader)
			s.handleGetAllUserTweets().ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}
