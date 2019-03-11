package logger

import (
	"net"
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// Logger handles all secure service logging
type Logger interface {
	LogRequest(*http.Request, string, string, int, string)
	LogDBRequest(string, string, ...interface{})

	LogError(error, string)
	LogStart(string)
}

func (l *logger) LogStart(port string) {
	l.out.WithFields(log.Fields{
		"transport": "http",
		"port":      port,
		"server":    l.server,
		"env":       l.env,
		"version":   l.version,
	}).Info("listening")
}

func (l *logger) LogError(err error, id string) {
	l.out.WithField("user_id", id).Error(err.Error())
}

func (l *logger) LogRequest(r *http.Request, duration, errString string, code int, id string) {
	username := "-"
	if r.URL.User != nil {
		if name := r.URL.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		host = r.RemoteAddr
	}

	uri := r.RequestURI

	if r.ProtoMajor == 2 && r.Method == "CONNECT" {
		uri = r.Host
	}

	if uri == "" {
		uri = r.URL.RequestURI()
	}

	l.out.WithFields(log.Fields{
		"type":     "request",
		"host":     host,
		"uri":      uri,
		"method":   r.Method,
		"username": username,
		"duration": duration,
		"agent":    r.Header.Get("User-Agent"),
		"status":   code,
		"user_id":  id,
		"server":   l.server,
		"env":      l.env,
		"version":  l.version,
	}).Info(errString)
}

// LogDBRequest logs the sql command
func (l *logger) LogDBRequest(cmd string, id string, args ...interface{}) {
	fields := log.Fields{
		"type":    "db",
		"user_id": id,
		"server":  l.server,
		"env":     l.env,
		"version": l.version,
	}
	for i, v := range args {
		fields["$"+strconv.Itoa(i+1)] = v
	}
	l.out.WithFields(fields).Print(cmd)
}

type logger struct {
	out     *log.Logger
	server  string
	env     string
	version string
}

// NewLogger instantiates the secure service logger
func NewLogger(server, env, version string) Logger {
	l := log.New()
	l.SetFormatter(&log.JSONFormatter{})
	l.SetOutput(os.Stdout)
	return &logger{l, server, env, version}
}
