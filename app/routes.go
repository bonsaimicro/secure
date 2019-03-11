package app

import (
	"encoding/json"
	"net/http"
	"secure/database"
)

type hFunc func(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus

var protectedRoutes = []struct {
	key string
	H   hFunc
}{
	{"/", proxy},
}

var openRoutes = []struct {
	key string
	H   hFunc
}{
	{"/healthz", health},
	{"/signup", signup},
	{"/login", login},
}

// Error is the handler's error interface
type Error interface {
	error
	Status() int
}

// httpStatus represents an error with an associated HTTP status code.
type httpStatus struct {
	Code          int
	res           []byte
	ResponseError string
	FuncErr       error
}

// httpAllows Status to satisfy the error interface.
func (se httpStatus) Error() string {
	return se.ResponseError
}

// Status returns our HTTP status code.
func (se httpStatus) Status() int {
	return se.Code
}

// Handler is the handler that holds a modified handler func
// and a pointer to the Env var
type Handler struct {
	*Env
	H hFunc
}

type context struct {
	User database.User
}

type results struct {
	Status  string      `json:"status"`
	Results interface{} `json:"results"`
}

type result struct {
	Status  string      `json:"status"`
	Results interface{} `json:"result"`
}

// ServeHTTP allows our Handler type to satisfy http.Handler interface
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := h.H(h.Env, w, r, &context{})
	if status.ResponseError != "" {
		http.Error(w, `{"status":"failure","result":"`+status.Error()+`"}`, status.Status())
	}
	w.Write(status.res)
}

// SetupRoutes sets up the routes
func (e *Env) setupRoutes() {
	for _, f := range protectedRoutes {
		r.Handle(f.key, setupMiddleware(Handler{e, f.H}))
	}
	for _, f := range openRoutes {
		r.Handle(f.key, logRequests(addToken(Handler{e, f.H})))
	}
}

func health(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
	return httpStatus{http.StatusOK, []byte(`{"status":"OK"}`), "", nil}
}

func login(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
	if r.Method == http.MethodPost {
		var a struct {
			Email    string
			Password string
		}

		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			return httpStatus{http.StatusBadRequest, nil, "Could not decode request body", err}
		}

		if a.Email == "" || a.Password == "" {
			return httpStatus{http.StatusBadRequest, nil, "Please supply an email and password", nil}
		}

		user, err := e.db.FindUser(a.Email, a.Password)
		c.User = database.User{Email: user.Email, FirstName: user.FirstName, LastName: user.LastName}

		if err != nil {
			return httpStatus{http.StatusInternalServerError, nil, err.Error(), err}
		}

		res, err := json.Marshal(result{"success", user})

		if err != nil {
			return httpStatus{http.StatusInternalServerError, nil, "Error marshalling results", err}
		}

		return httpStatus{http.StatusOK, res, "", nil}
	}
	return httpStatus{http.StatusNotFound, nil, http.StatusText(http.StatusNotFound), nil}
}

func signup(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
	if r.Method == http.MethodPost {
		var a struct {
			Email    string
			Password string
		}

		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			return httpStatus{http.StatusBadRequest, nil, "Could not decode request body", err}
		}

		if a.Email == "" || a.Password == "" {
			return httpStatus{http.StatusBadRequest, nil, "Please supply an email and password", nil}
		}

		user, err := database.NewUser(a.Email, a.Password)
		c.User = database.User{Email: user.Email, FirstName: user.FirstName, LastName: user.LastName}

		if err != nil {
			return httpStatus{http.StatusInternalServerError, nil, "", err}
		}

		user, err = e.db.AddUser(user)

		if err != nil {
			return httpStatus{http.StatusBadRequest, nil, err.Error(), err}
		}

		res, err := json.Marshal(result{"success", user})

		if err != nil {
			return httpStatus{http.StatusInternalServerError, nil, "Error marshalling results", err}
		}

		return httpStatus{http.StatusOK, res, "", nil}
	}
	return httpStatus{http.StatusNotFound, nil, http.StatusText(http.StatusNotFound), nil}
}

func proxy(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
	res, err := json.Marshal(result{"success", c.User})

	if err != nil {
		return httpStatus{http.StatusInternalServerError, nil, "Error marshalling results", err}
	}

	return httpStatus{http.StatusOK, res, "", nil}
}
