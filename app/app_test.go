package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"secure/database"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
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

func TestMain(m *testing.M) {
	patch := monkey.Patch(time.Now, func() time.Time {
		return time.Date(2019, 1, 1, 1, 1, 1, 1, time.UTC)
	})
	defer patch.Unpatch()

	os.Setenv("DARE_PASSWORD", "test")
	os.Setenv("DARE_SALT", "test")
	os.Setenv("FORWARD_URL", "www.google.com")
	db, err := database.New("/tmp/badger_test_db", &loggerX{})
	if err != nil {
		panic(err)
	}
	env := Env{db, &loggerX{}, nil}
	env.setupRoutes()
	m.Run()
	db.Close()
	os.RemoveAll("/tmp/badger_test_db")
}

func TestSignUp(t *testing.T) {
	email := uniuri.New() + "@me.com"
	pass := uniuri.New()

	payload := []byte(`{"email":"` + email + `","password":"` + pass + `"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(payload))
	response := executeRequest(req)
	assert.Equal(t, response.Code, http.StatusOK, "Response should be 200")
	user, _ := database.NewUser(email, pass)
	user.PasswordSalt = []byte{}
	res := struct {
		Status string         `json:"status"`
		Result *database.User `json:"result"`
	}{"success", user}
	assert.Equal(t, response.Body.String(), structToString(res), "Response body should match")
}

func TestLogin(t *testing.T) {
	email := uniuri.New() + "@me.com"
	pass := uniuri.New()

	payload := []byte(`{"email":"` + email + `","password":"` + pass + `"}`)
	signup, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(payload))
	executeRequest(signup)
	login, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(login)
	assert.Equal(t, response.Code, http.StatusOK, "Response should be 200")
	user, _ := database.NewUser(email, pass)
	user.PasswordSalt = []byte{}

	res := struct {
		Status string         `json:"status"`
		Result *database.User `json:"result"`
	}{"success", user}
	assert.Equal(t, response.Body.String(), structToString(res), "Response body should match")
}

func TestCanCallAuthenticatedEndPoint(t *testing.T) {
	email := uniuri.New() + "@me.com"
	pass := uniuri.New()

	payload := []byte(`{"email":"` + email + `","password":"` + pass + `"}`)
	signup, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(payload))
	resp := executeRequest(signup)

	proxy, _ := http.NewRequest("POST", "/proxy", bytes.NewBuffer(payload))
	proxy.Header.Set("some", "header")
	proxy.AddCookie(resp.Result().Cookies()[0])
	response := executeRequest(proxy)
	assert.Equal(t, response.Code, http.StatusOK, "Response should be 200")
}
