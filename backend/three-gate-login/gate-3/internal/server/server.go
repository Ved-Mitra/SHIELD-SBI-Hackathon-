package server

import (
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"golang.org/x/time/rate"

	"shield/gate3/internal/config"
	"shield/gate3/internal/handler"
	"shield/gate3/internal/middleware"
	"shield/gate3/internal/store"
)

func New(cfg config.Config) (http.Handler, error) {
	wa, err := webauthn.New(cfg.WebAuthn)
	if err != nil {
		return nil, err
	}

	sessions := store.NewMemorySessionStore()
	userStore := store.NewInMemoryUserStore()

	webAuthnHandler := &handler.WebAuthnHandler{
		WebAuthn:  wa,
		Sessions:  sessions,
		UserStore: userStore,
	}

	rl := middleware.NewRateLimiter(rate.Every(3*time.Second), 20)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.Health)

	// FIDO2 endpoints
	fidoMux := http.NewServeMux()
	fidoMux.HandleFunc("/gate3/register/begin", webAuthnHandler.RegisterBegin)
	fidoMux.HandleFunc("/gate3/register/finish", webAuthnHandler.RegisterFinish)
	fidoMux.HandleFunc("/gate3/authenticate/begin", webAuthnHandler.AuthBegin)
	fidoMux.HandleFunc("/gate3/authenticate/finish", webAuthnHandler.AuthFinish)

	var authChain http.Handler = fidoMux
	authChain = middleware.Gate2Auth(cfg.Gate2PublicKey, authChain)
	authChain = rl.Middleware(authChain)

	mux.Handle("/gate3/", authChain)

	var h http.Handler = mux
	h = middleware.RequestID(h)
	h = middleware.Logging(h)

	return h, nil
}

func DefaultServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
