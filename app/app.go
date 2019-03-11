package app

import (
	"net/http"
	"os"
	"secure/database"
	"secure/logger"
	"secure/server"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

var (
	r       = http.NewServeMux()
	version = "0.0.1"
)

// Env sets the environment variables for the application
type Env struct {
	db  database.Datastore
	l   logger.Logger
	jwt *jwtmiddleware.JWTMiddleware
}

// Start starts the application
func Start() {
	l := logger.NewLogger("secure", os.Getenv("ENV"), version)

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("My Secret"), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	db, err := database.New(os.Getenv("DB_PATH"), l)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	env := Env{db, l, jwtMiddleware}
	env.setupRoutes()

	server.New(l, r)
}
