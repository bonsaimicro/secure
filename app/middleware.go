package app

import (
	"net/http"
	"os"
	"secure/database"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	User database.User `json:"user"`
	jwt.StandardClaims
}

func setupMiddleware(r Handler) Handler {
	return logRequests(checkToken(r))
}

func logRequests(h Handler) Handler {
	return Handler{h.Env, hFunc(func(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
		t := time.Now()
		res := h.H(h.Env, w, r, c)
		go e.l.LogRequest(r, time.Since(t).String(), res.Error(), res.Status(), c.User.Email)
		if res.FuncErr != nil {
			go e.l.LogError(res.FuncErr, c.User.Email)
		}
		return res
	})}
}

func checkToken(h Handler) Handler {
	return Handler{h.Env, hFunc(func(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
		var tokenString string
		tokens, ok := r.Header["Authorization"]
		if ok && len(tokens) >= 1 {
			tokenString = tokens[0]
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		if tokenString == "" {
			cookie, err := r.Cookie("jwt")
			if err != nil {
				return httpStatus{http.StatusUnauthorized, nil, http.StatusText(http.StatusUnauthorized), err}
			}
			tokenString = cookie.String()
			tokenString = strings.TrimPrefix(tokenString, "jwt=")
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			return httpStatus{http.StatusUnauthorized, nil, http.StatusText(http.StatusUnauthorized), err}
		}

		if !token.Valid {
			if ve, okVal := err.(*jwt.ValidationError); okVal {
				return httpStatus{http.StatusUnauthorized, nil, ve.Error(), nil}
			}
			return httpStatus{http.StatusUnauthorized, nil, http.StatusText(http.StatusUnauthorized), nil}
		}

		if claims, ok := token.Claims.(*Claims); ok {
			c.User = claims.User
		}

		return h.H(h.Env, w, r, c)
	})}
}

func addToken(h Handler) Handler {
	return Handler{h.Env, hFunc(func(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
		res := h.H(h.Env, w, r, c)
		if res.FuncErr != nil {
			time.Sleep(100 * time.Millisecond)
			return res
		}

		expires := time.Now().Local().Add(48 * time.Hour)
		mySigningKey := []byte(os.Getenv("JWT_SECRET"))

		claims := Claims{
			c.User,
			jwt.StandardClaims{
				ExpiresAt: expires.Unix(),
				Issuer:    "server",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
		ss, signErr := token.SignedString(mySigningKey)

		if signErr != nil {
			return httpStatus{http.StatusInternalServerError, nil, "", signErr}
		}

		cookie := http.Cookie{Name: "jwt", Value: ss, Expires: expires}
		http.SetCookie(w, &cookie)
		return res
	})}
}
