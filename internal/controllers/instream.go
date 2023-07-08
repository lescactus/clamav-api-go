package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog/hlog"
)

// InStreamResponse represents the json response of a /scan endpoint.
type InStreamResponse struct {
	Status     string `json:"status"`
	Msg        string `json:"msg"`
	Signature  string `json:"signature"`
	VirusFound bool   `json:"virus_found"`
}

var (
	ErrFormFile        = errors.New("failed to parse file")
	ErrOpenFileHeaders = errors.New("failed to open multipart file headers")
)

func (h *Handler) InStream(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	// Parsing the Multipart file
	_, hd, err := r.FormFile("file")
	if err != nil {
		e := fmt.Errorf("%w: %w", ErrFormFile, err)
		h.Logger.Debug().Str("req_id", req_id.String()).Msgf("%v", e)

		SetErrorResponse(w, e)
		return
	}

	f, err := hd.Open()
	if err != nil {
		e := fmt.Errorf("%w: %w", ErrOpenFileHeaders, err)
		h.Logger.Debug().Str("req_id", req_id.String()).Msgf("%v", e)

		SetErrorResponse(w, e)
		return
	}

	defer f.Close()

	size := hd.Size

	h.Logger.Debug().
		Str("req_id", req_id.String()).
		Str("file_name", hd.Filename).
		Int64("file_size", hd.Size).
		Msg("multipart file read successfully")

	var inStreamResp InStreamResponse
	var ctx = r.Context()

	inStream, err := h.Clamav.InStream(ctx, f, size)
	if err != nil {
		if errors.Is(err, clamav.ErrVirusFound) {
			h.Logger.Debug().Str("req_id", req_id.String()).Msg(err.Error())

			inStreamResp = InStreamResponse{
				Status:     "error",
				Msg:        clamav.ErrVirusFound.Error(),
				Signature:  h.parseSignature(string(inStream)),
				VirusFound: true,
			}
		} else {
			h.Logger.Debug().Str("req_id", req_id.String()).Err(err).Msg("error while scanning file")

			SetErrorResponse(w, err)
			return
		}
	} else {
		inStreamResp = InStreamResponse{
			Status:     "noerror",
			Msg:        string(clamav.RespScan),
			Signature:  "",
			VirusFound: false,
		}
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("file scanned successfully")

	resp, err := json.Marshal(inStreamResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// parseSignature will extract the name of the virus signature
// from Clamd response when a potential virus is found.
//
// An example of such response from the Clamd daemon is:
// "stream: Eicar-Signature FOUND"
func (h *Handler) parseSignature(msg string) string {
	return strings.TrimLeft(strings.TrimRight(msg, " FOUND"), "stream: ")
}
