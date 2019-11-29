package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/google/uuid"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store"
)

type server struct {
	router *mux.Router
	logger *logrus.Logger
	store  store.Store
}

const (
	ctxKeyRequestID int8   = iota
)

var (
	jwtKey                      = []byte("my_secret_key")
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated         = errors.New("not authenticated")
)

type Claims struct {
	AccountId int    `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
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

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.HandleFunc("/register", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/login", s.handleUsersLogin()).Methods("POST")
	s.router.HandleFunc("/tweets", s.handleTweetsCreate()).Methods("POST")
	s.router.HandleFunc("/tweets", s.handleGetAllTweetsFromSubscriptions()).Methods("GET")
	s.router.HandleFunc("/mytweets", s.handleGetAllUserTweets()).Methods("GET")
	s.router.HandleFunc("/subscribe", s.handleSubscribeToUser()).Methods("POST")
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

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

		// init jwt token
		tokenString, err := s.authenticateUserWithJwt(w, r, u)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, JwtToken{TokenString: tokenString})
	}
}

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

		userId, err, code := s.checkAuthenticateUserWithJwt(w, r)
		if err != nil {
			s.error(w, r, code, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			// TODO: set better error message for this case
			s.error(w, r, http.StatusInternalServerError, errors.New("cant find user in db"))
			return
		}

		su, err := s.store.User().FindByUsername(req.Nickname)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, errors.New("cant find user in db"))
			return
		}

		if err := s.store.User().SubscribeTo(u, su); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, nil)
	}
}

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

		userId, err, code := s.checkAuthenticateUserWithJwt(w, r)
		if err != nil {
			s.error(w, r, code, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			// TODO: set better error message for this case
			s.error(w, r, http.StatusInternalServerError, errors.New("cant find user in db"))
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

func (s *server) handleGetAllTweetsFromSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err, code := s.checkAuthenticateUserWithJwt(w, r)
		if err != nil {
			s.error(w, r, code, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			// TODO: set better error message for this case
			s.error(w, r, http.StatusInternalServerError, errors.New("cant find user in db"))
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

func (s *server) handleGetAllUserTweets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err, code := s.checkAuthenticateUserWithJwt(w, r)
		if err != nil {
			s.error(w, r, code, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(userId)
		if err != nil {
			// TODO: set better error message for this case
			s.error(w, r, http.StatusInternalServerError, errors.New("cant find user in db"))
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

func (s *server) checkAuthenticateUserWithJwt(w http.ResponseWriter, r *http.Request) (int, error, int) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			return 0, err, http.StatusUnauthorized
		}
		// For any other type of error, return a bad request status
		return 0, err, http.StatusBadRequest
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return 0, err, http.StatusUnauthorized
		}
		return 0, err, http.StatusBadRequest
	}
	if !tkn.Valid {
		return 0, err, http.StatusUnauthorized
	}

	return claims.AccountId, nil, http.StatusOK
}

func (s *server) authenticateUserWithJwt(w http.ResponseWriter, r *http.Request, u *model.User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		AccountId: u.Id,
		Email:     u.Email,
		Username:  u.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}
	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	return tokenString, nil
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
