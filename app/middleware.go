package app

import (
	"net/http"
	"time"
)

func setupMiddleware(r Handler) Handler {
	return logRequests(authenticate(r))
}

func logRequests(h Handler) Handler {
	return Handler{h.Env, hFunc(func(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
		t := time.Now()
		err := h.H(h.Env, w, r, c)
		go e.l.LogRequest(r, time.Since(t).String(), err.Error(), err.Status(), c.Email)
		if err.FuncErr != nil {
			go e.l.LogError(err.FuncErr, c.Email)
		}
		return err
	})}
}

func authenticate(h Handler) Handler {
	return Handler{h.Env, hFunc(func(e *Env, w http.ResponseWriter, r *http.Request, c *context) httpStatus {
		return h.H(h.Env, w, r, c)
	})}
}
