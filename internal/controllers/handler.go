package controllers

import (
	"net/http"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog"
)

const (
	// ContentTypeApplicationJSON represent the applcation/json Content-Type value
	ContentTypeApplicationJSON = "application/json"
)

type Handler struct {
	Clamav clamav.Clamaver
	Logger *zerolog.Logger
}

func NewHandler(logger *zerolog.Logger, clamav clamav.Clamaver) *Handler {
	return &Handler{Logger: logger, Clamav: clamav}
}

// MaxReqSizeis a HTTP middleware limiting the size of the request
// by using http.MaxBytesReader() on the request body.
func MaxReqSize(maxReqSize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxReqSize)
			next.ServeHTTP(w, r)
		})
	}
}
