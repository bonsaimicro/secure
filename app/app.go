package app

import (
	"net/http"
	"os"
	"secure/database"
	"secure/logger"
	"secure/server"
)

var (
	r       = http.NewServeMux()
	version = "0.0.1"
)

// Env sets the environment variables for the application
type Env struct {
	db database.Datastore
	l  logger.Logger
}

// Start starts the application
func Start() {
	l := logger.NewLogger("secure", os.Getenv("ENV"), version)

	db, err := database.New(os.Getenv("DB_PATH"), l)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	env := Env{db, l}
	env.setupRoutes()

	server.New(l, r)
}
