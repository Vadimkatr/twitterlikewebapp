package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
)

type server struct {
	router *mux.Router
	logger *logrus.Logger
	store  store.Store
}

const (
	jwtAccessExpTimeMin  time.Duration = 5
	jwtRefreshExpTimeMin time.Duration = 10
	ctxKeyRequestID      int8          = iota
)

var (
	jwtKey                      = []byte("my_secret_key_that_will_be_very_secret")
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
)

type accessClaims struct {
	UserId   int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.StandardClaims
}

type refreshClaims struct {
	jwt.StandardClaims
}

type JwtToken struct {
	TokenString string `json:"token"`
}

func newServer(store store.Store) *server {
	s := &server{
		router: mux.NewRouter(),
		logger: logrus.New(),
		store:  store,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, s.router)
	loggedRouter.ServeHTTP(w, r)
}

// configureRouter - set server routes
func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.HandleFunc("/register", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/login", s.handleUsersLogin()).Methods("POST")
	s.router.HandleFunc("/tweets", s.authMiddleware(s.handleTweetsCreate())).Methods("POST")
	s.router.HandleFunc("/tweets", s.authMiddleware(s.handleGetAllTweetsFromSubscriptions())).Methods("GET")
	s.router.HandleFunc("/mytweets", s.authMiddleware(s.handleGetAllUserTweets())).Methods("GET")
	s.router.HandleFunc("/subscribe", s.authMiddleware(s.handleSubscribeToUser())).Methods("POST")
}

// setRequestID - middleware that add special uuid for request
func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

// handleUsersCreate - create user in db
func (s *server) handleUsersCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email:    req.Email,
			Username: req.Username,
			Password: req.Password,
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}
}

// handleUsersLogin - try to find user in db and respond jwt to user
func (s *server) handleUsersLogin() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		// Init jwt
		tokenString, err := s.authenticate(w, r, u)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, JwtToken{TokenString: tokenString})
	}
}

// handleSubscribeToUser - add note to db that user subscribe to another user
func (s *server) handleSubscribeToUser() http.HandlerFunc {
	type request struct {
		Nickname string `json:"nickname"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		userIdStr := r.Header.Get("user_id")
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, nil)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			if err == store.ErrUserNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrUserNotFound)
			} else {
				s.error(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		su, err := s.store.User().FindByUsername(req.Nickname)
		if err != nil {
			if err == store.ErrUserNotFound {
				s.error(w, r, http.StatusBadRequest, store.ErrUserNotFound)
			} else {
				s.error(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		if err := s.store.User().SubscribeTo(u, su); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, nil)
	}
}

// handleTweetsCreate - add user tweet to db; respond is tweet id adn tweet message
func (s *server) handleTweetsCreate() http.HandlerFunc {
	type request struct {
		Message string `json:"message"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		userIdStr := r.Header.Get("user_id")
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, nil)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			if err == store.ErrUserNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrUserNotFound)
			} else {
				s.error(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		t := &model.Tweet{
			Message: req.Message,
			UserId:  u.Id,
		}
		if err := s.store.Tweet().Create(t); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, map[string]string{"id": string(t.Id), "message": t.Message})
	}
}

// handleGetAllTweetsFromSubscriptions - find all tweets of user subscribtions
func (s *server) handleGetAllTweetsFromSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIdStr := r.Header.Get("user_id")
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, nil)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			if err == store.ErrUserNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrUserNotFound)
			} else {
				s.error(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		tweets, err := s.store.Tweet().FindTweetsFromSubscriptions(u.Id)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusOK, map[string][]string{"tweets": tweets})
	}
}

// handleGetAllUserTweets - get all user tweets
func (s *server) handleGetAllUserTweets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIdStr := r.Header.Get("user_id")
		userId, err := strconv.Atoi(userIdStr)

		if err != nil {
			s.error(w, r, http.StatusInternalServerError, nil)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			if err == store.ErrUserNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrUserNotFound)
			} else {
				s.error(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		tweets, err := s.store.Tweet().GetAllUserTweets(u.Id)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusOK, map[string][]string{"tweets": tweets})
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
