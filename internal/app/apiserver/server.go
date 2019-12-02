package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
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
	errNotAuthenticated         = errors.New("not authenticated")
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
	s.router.ServeHTTP(w, r)
}

// configureRouter - set server routes
func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.HandleFunc("/register", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/login", s.handleUsersLogin()).Methods("POST")
	s.router.HandleFunc("/tweets", s.authmiddlewareWithJwt(s.handleTweetsCreate())).Methods("POST")
	s.router.HandleFunc("/tweets", s.authmiddlewareWithJwt(s.handleGetAllTweetsFromSubscriptions())).Methods("GET")
	s.router.HandleFunc("/mytweets", s.authmiddlewareWithJwt(s.handleGetAllUserTweets())).Methods("GET")
	s.router.HandleFunc("/subscribe", s.authmiddlewareWithJwt(s.handleSubscribeToUser())).Methods("POST")
}

// setRequestID - amiddleware that add special uuid for request
func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

// logRequest - logging middleware
func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
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
		tokenString, err := s.authenticateUserWithJwt(w, r, u)
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
			if err == store.ErrRecordNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrRecordNotFound)
			} else {
				s.error(w, r, http.StatusInternalServerError, err)
			}
			return
		}

		su, err := s.store.User().FindByUsername(req.Nickname)
		if err != nil {
			if err == store.ErrRecordNotFound {
				s.error(w, r, http.StatusBadRequest, store.ErrRecordNotFound)
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
			if err == store.ErrRecordNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrRecordNotFound)
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
			if err == store.ErrRecordNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrRecordNotFound)
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
		logrus.Println("AAAAAAAAAAA |", userIdStr, "|", userId,"|", err)

		if err != nil {
			s.error(w, r, http.StatusInternalServerError, nil)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			if err == store.ErrRecordNotFound {
				s.error(w, r, http.StatusInternalServerError, store.ErrRecordNotFound)
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

func (s *server) authmiddlewareWithJwt(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("access_token")
		if err != nil {
			if err == http.ErrNoCookie {
				s.error(w, r, http.StatusUnauthorized, err)
				return
			}
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		atString := c.Value
		atClaims := &accessClaims{}

		tkn, err := jwt.ParseWithClaims(atString, atClaims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				s.error(w, r, http.StatusUnauthorized, err)
				return
			}
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if time.Unix(atClaims.ExpiresAt, 0).Sub(time.Now()) > 30 * time.Second {
			atExpirationTime := time.Now().Add(jwtAccessExpTimeMin * time.Minute)
			atString, err = s.createToken(atExpirationTime, atClaims)
			if err != nil {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
		}
	
		if !tkn.Valid {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}
		
		userId := strconv.Itoa(atClaims.UserId)
		r.Header.Set("user_id", userId)
		next.ServeHTTP(w, r)
	})
}

func (s *server) authenticateUserWithJwt(w http.ResponseWriter, r *http.Request, u *model.User) (string, error) {
	// Accesse Token init
	atExpirationTime := time.Now().Add(1 * time.Minute)
	atClaims := &accessClaims{
		UserId:   u.Id,
		Username: u.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: atExpirationTime.Unix(),
		},
	}
	atString, err := s.createToken(atExpirationTime, atClaims)
	if err != nil {
		return "", err
	}

	// Refresh Token init
	rtExpirationTime := time.Now().Add(10 * time.Minute)
	rtClaims := &refreshClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: rtExpirationTime.Unix(),
		},
	}
	rtString, err :=s.createToken(rtExpirationTime, rtClaims)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "access_token",
		Value:   atString,
		Expires: atExpirationTime,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh_token",
		Value:   rtString,
		Expires: rtExpirationTime,
	})

	return atString, nil
}

func (s *server) createToken(expTime time.Time, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, err
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
