package controllers

import (
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
