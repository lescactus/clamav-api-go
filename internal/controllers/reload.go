package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog/hlog"
)

// ReloadResponse represents the json response of a /reload endpoint.
type ReloadResponse struct {
	Status string `json:"status"`
}

func (h *Handler) Reload(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	err := h.Clamav.Reload(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msgf("error while sending version command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("version command sent successfully")

	reload := ReloadResponse{
		Status: string(clamav.RespReload),
	}

	resp, err := json.Marshal(&reload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
