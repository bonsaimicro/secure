package server

import (
	"secure/logger"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// New creates a new TLS encrpted https server
func New(l logger.Logger, mux *http.ServeMux) {
	if os.Getenv("ENV") == "production" {
		dataDir := "."
		hostPolicy := func(ctx context.Context, host string) error {
			// Note: change to your real domain
			allowedHost := os.Getenv("DOMAIN")
			if host == allowedHost {
				return nil
			}
			return fmt.Errorf("acme/autocert: only %s host is allowed", allowedHost)
		}

		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(dataDir),
		}

		cnf := tls.Config{
			GetCertificate: m.GetCertificate,
			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
		}

		srv := &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      mux,
			TLSConfig:    &cnf,
		}
		go l.LogStart(":443")
		log.Fatal(srv.ListenAndServeTLS("", ""))
	} else {
		port := ":" + os.Getenv("PORT")
		go l.LogStart(port)
		log.Fatal(http.ListenAndServe(port, mux))
	}
}
