package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// PingResponse represents the json response of a /ping endpoint
type PingResponse struct {
	Ping string `json:"ping"`
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	ping, err := h.Clamav.Ping(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msgf("error while sending ping command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("ping command sent successfully")

	p := PingResponse{
		Ping: string(ping),
	}

	resp, err := json.Marshal(&p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
