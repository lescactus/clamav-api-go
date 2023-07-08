package main

import (
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/lescactus/clamav-api-go/internal/config"
	"github.com/lescactus/clamav-api-go/internal/controllers"
	"github.com/lescactus/clamav-api-go/internal/logger"
	"github.com/rs/zerolog/hlog"
)

func main() {
	// Get application configuration
	cfg := config.New()

	logger := logger.New(
		cfg.GetString("LOGGER_LOG_LEVEL"),
		cfg.GetString("LOGGER_DURATION_FIELD_UNIT"),
		cfg.GetString("LOGGER_FORMAT"),
	)

	client := clamav.NewClamavClient(
		cfg.GetString("CLAMAV_ADDR"),
		cfg.GetString("CLAMAV_NETWORK"),
		cfg.GetDuration("CLAMAV_TIMEOUT"),
		cfg.GetDuration("CLAMAV_KEEPALIVE"),
	)

	// Create http router, server and handler controller
	r := httprouter.New()
	h := controllers.NewHandler(logger, client)
	c := alice.New()
	s := &http.Server{
		Addr:              cfg.GetString("APP_ADDR"),
		Handler:           handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r), // recover from panics and print recovery stack
		ReadTimeout:       cfg.GetDuration("SERVER_READ_TIMEOUT"),
		ReadHeaderTimeout: cfg.GetDuration("SERVER_READ_HEADER_TIMEOUT"),
		WriteTimeout:      cfg.GetDuration("SERVER_WRITE_TIMEOUT"),
	}

	// logger fields
	*logger = logger.With().Str("svc", config.AppName).Logger()

	// // Register logging middleware
	c = c.Append(hlog.NewHandler(*logger))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RemoteAddrHandler("remote_client"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RequestIDHandler("req_id", "X-Request-ID"))

	r.Handler(http.MethodGet, "/rest/v1/ping", c.ThenFunc(h.Ping))
	r.Handler(http.MethodGet, "/rest/v1/version", c.ThenFunc(h.Version))
	r.Handler(http.MethodGet, "/rest/v1/stats", c.ThenFunc(h.Stats))
	r.Handler(http.MethodGet, "/rest/v1/versioncommands", c.ThenFunc(h.VersionCommands))
	r.Handler(http.MethodPost, "/rest/v1/reload", c.ThenFunc(h.Reload))
	r.Handler(http.MethodPost, "/rest/v1/shutdown", c.ThenFunc(h.Shutdown))
	r.Handler(http.MethodPost, "/rest/v1/scan", c.ThenFunc(h.InStream))

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Startup failed")
	}
}
