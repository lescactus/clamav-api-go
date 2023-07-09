package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof" // Register the pprof handlers

	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/lescactus/clamav-api-go/internal/config"
	"github.com/lescactus/clamav-api-go/internal/controllers"
	"github.com/lescactus/clamav-api-go/internal/logger"
	"github.com/rs/zerolog/hlog"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func main() {
	// Get application configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("unable to build a new app config: %v", err)
	}

	logger := logger.New(
		cfg.LoggerLogLevel,
		cfg.LoggerDurationFieldUnit,
		cfg.LoggerFormat,
	)

	client := clamav.NewClamavClient(
		cfg.ClamavAddr,
		cfg.ClamavNetwork,
		cfg.ClamavTimeout,
		cfg.ClamavKeepAlive,
	)

	// Create http router, server and handler controller
	r := httprouter.New()
	h := controllers.NewHandler(logger, client)
	c := alice.New()
	s := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r), // recover from panics and print recovery stack
		ReadTimeout:       cfg.ServerReadTimeout,
		ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
		WriteTimeout:      cfg.ServerWriteTimeout,
	}

	// logger fields
	*logger = logger.With().Str("svc", config.AppName).Logger()

	// Register logging middleware
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

	// Start server
	go func() {
		logger.Info().Msgf("Starting server %s on address %s ...", config.AppName, cfg.ServerAddr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Startup failed")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Blocking until receiving a shutdown signal
	sig := <-sigChan

	logger.Info().Msgf("Server received %s signal. Shutting down...", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	// Attempting to gracefully shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		logger.Warn().Msg("Failed to gracefully shutdown the server")
	}

}
