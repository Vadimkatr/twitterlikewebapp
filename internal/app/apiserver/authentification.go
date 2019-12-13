package apiserver

import (
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/model"
)

var (
	jwtAccessExpTimeMin  time.Duration = 5 * time.Minute
	jwtRefreshExpTimeMin time.Duration = 20 * time.Minute
	expiryTimeToUpdateAT time.Duration = 30 * time.Second
)

func (s *server) authMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get access_token from cookies
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

		// Parse JWT and validate them
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
		if !tkn.Valid {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		// New token will only be issued if the old token is within expiryTimeToUpdateAT seconds of expiry.
		if time.Unix(atClaims.ExpiresAt, 0).Sub(time.Now()) < 30*time.Second {
			token, tokenExpTime, err, code := s.updateAccessToken(w, r, atClaims)
			if err != nil {
				s.error(w, r, code, err)
			}
			// Set new tokens to cookies
			http.SetCookie(w, &http.Cookie{
				Name:    "access_token",
				Value:   token,
				Expires: tokenExpTime,
			})
		}

		userId := strconv.Itoa(atClaims.UserId)
		r.Header.Set("user_id", userId)
		next.ServeHTTP(w, r)
	})
}

// authenticate - create access and refresh token for user after login and set thems to cookies
func (s *server) authenticate(w http.ResponseWriter, r *http.Request, u *model.User) (string, error) {
	// Access Token init
	atExpirationTime := time.Now().Add(jwtAccessExpTimeMin)
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
	rtExpirationTime := time.Now().Add(jwtRefreshExpTimeMin)
	rtClaims := &refreshClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: rtExpirationTime.Unix(),
		},
	}
	rtString, err := s.createToken(rtExpirationTime, rtClaims)
	if err != nil {
		return "", err
	}

	// Set tokens to cookies
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

// updateAccessToken - get refresh_token, validate them, and if its valid => create new access_token
func (s *server) updateAccessToken(w http.ResponseWriter, r *http.Request, atOldClaims *accessClaims) (string, time.Time, error, int) {
	// Get refresh_token from cookies
	c, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return "", time.Time{}, err, http.StatusUnauthorized
		}
		return "", time.Time{}, err, http.StatusBadRequest
	}

	rtString := c.Value
	rtClaims := &refreshClaims{}

	// Parse refresh_token and validate them
	tkn, err := jwt.ParseWithClaims(rtString, rtClaims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", time.Time{}, err, http.StatusUnauthorized
		}
		return "", time.Time{}, err, http.StatusBadRequest
	}
	// If token isnt valid => them response AuthError and StatusUnauthorized
	if !tkn.Valid {
		return "", time.Time{}, ErrNotAuthenticated, http.StatusUnauthorized
	}

	// Create new access_token with old claims (only update exp time)
	atExpirationTime := time.Now().Add(jwtAccessExpTimeMin)
	atOldClaims.ExpiresAt = atExpirationTime.Unix()
	atString, err := s.createToken(atExpirationTime, atOldClaims)
	if err != nil {
		return "", time.Time{}, err, http.StatusInternalServerError
	}

	return atString, atExpirationTime, nil, http.StatusOK
}

// Create new jwt with claims (its can be access or refresh token)
func (s *server) createToken(expTime time.Time, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, err
}
