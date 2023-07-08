package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// ShutdownResponse represents the json response of a /shutdown endpoint.
type ShutdownResponse struct {
	Status string `json:"status"`
}

func (h *Handler) Shutdown(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	err := h.Clamav.Shutdown(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msgf("error while sending shutdown command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("shutdown command sent successfully")

	shutdown := ShutdownResponse{
		Status: "Shutting down",
	}

	resp, err := json.Marshal(&shutdown)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
