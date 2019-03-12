package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

		if err != nil {
			return httpStatus{http.StatusInternalServerError, nil, err.Error(), err}
		}

		c.User = database.User{Email: user.Email, FirstName: user.FirstName, LastName: user.LastName}

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

		if err != nil {
			return httpStatus{http.StatusInternalServerError, nil, "", err}
		}

		c.User = database.User{Email: user.Email, FirstName: user.FirstName, LastName: user.LastName}

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
	res, err := json.Marshal(c.User)

	if err != nil {
		return httpStatus{http.StatusInternalServerError, nil, "Error marshalling results", err}
	}

	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return httpStatus{http.StatusBadRequest, nil, "Could not decode request body", err}
	}

	// you can reassign the body if you need to parse it as multipart
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("http://%s%s", os.Getenv("FORWARD_URL"), r.RequestURI)

	proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = r.Header
	proxyReq.Header = make(http.Header)
	for h, val := range r.Header {
		if h != "Cookie" {
			proxyReq.Header[h] = val
		}
	}

	proxyReq.Header.Set("X-Forwarded-User", string(res))

	httpClient := http.Client{}
	proxyResp, err := httpClient.Do(proxyReq)
	if err != nil {
		return httpStatus{http.StatusBadGateway, nil, http.StatusText(http.StatusBadGateway), err}
	}
	defer proxyResp.Body.Close()

	proxyBody, err := ioutil.ReadAll(proxyResp.Body)

	if err != nil {
		return httpStatus{http.StatusBadRequest, nil, "Could not decode request body of the proxied request", err}
	}

	return httpStatus{proxyResp.StatusCode, proxyBody, "", err}
}
