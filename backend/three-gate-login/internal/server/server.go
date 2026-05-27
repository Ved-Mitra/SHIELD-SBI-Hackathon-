package server

import (
	"net/http"

	"shield/three-gate-login/internal/config"
	"shield/three-gate-login/internal/handler"
	"shield/three-gate-login/internal/middleware"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.Health)
	mux.Handle("/gate2/token", handler.TokenHandler{Config: cfg})

	h := middleware.RequestID(mux)
	h = middleware.Logging(h)
	return h
}

