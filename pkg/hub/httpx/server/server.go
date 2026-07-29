package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"hara.sh/alloy/httpx/handlerx"
	"hara.sh/alloy/httpx/routing"
)

// Config controls the net/http server created for an Alloy router.
type Config struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ErrorLog          *log.Logger
	BaseContext       func(net.Listener) context.Context
}

// NewHandler returns an http.Handler that dispatches requests through router.
func NewHandler(router *routing.Router) http.Handler {
	return handlerx.New(router)
}

// New returns a configured net/http server for router.
func New(router *routing.Router, cfg Config) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(router),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ErrorLog:          cfg.ErrorLog,
		BaseContext:       cfg.BaseContext,
	}
}
