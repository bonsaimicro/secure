package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	datastore_mock "secure/database/mocks"
	"testing"
)

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func structToString(s interface{}) string {
	d, _ := json.Marshal(s)
	return string(d)
}

type loggerX struct{}

func (l *loggerX) LogRequest(*http.Request, string, string, int, string) {}
func (l *loggerX) LogDBRequest(string, string, ...interface{})           {}
func (l *loggerX) LogError(error, string)                                {}
func (l *loggerX) LogStart(string)                                       {}

var d datastore_mock.Datastore

func TestMain(m *testing.M) {
	d = datastore_mock.Datastore{}
	env := Env{&d, &loggerX{}, nil}
	env.setupRoutes()
	m.Run()
}
