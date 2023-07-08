package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/hlog"
)

// VersionCommandsResponse represents the json response of a /versioncommands endpoint.
// It represents the version of Clamav, followed by "| COMMANDS:" and a
// space-delimited list of supported commands.
type VersionCommandsResponse struct {
	Version  string   `json:"clamav_version"`
	Commands []string `json:"commands"`
}

func (h *Handler) VersionCommands(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	vcmds, err := h.Clamav.VersionCommands(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msgf("error while sending versioncommands command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("versioncommands command sent successfully")

	v, err := versionCommandsMarshall(string(vcmds))
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msgf("error while marshalling versioncommands: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("versioncommands marshalled successfully")

	resp, err := json.Marshal(&v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// versionCommandsMarshall will marshall the string v
// into a *VersionCommandsResponse.
// It returns an error if not possible.
func versionCommandsMarshall(v string) (*VersionCommandsResponse, error) {
	// The Clamav "VERSIONCOMMANDS" command
	// return the version of Clamav, followed by "| COMMANDS:" and a
	// space-delimited list of supported commands.
	//
	// ex: "ClamAV 1.0.0/26804/Mon Feb  6 08:47:07 2023| COMMANDS: SCAN QUIT RELOAD PING CONTSCAN VERSIONCOMMANDS VERSION END SHUTDOWN MULTISCAN FILDES STATS IDSESSION INSTREAM DETSTATSCLEAR DETSTATS ALLMATCHSCAN"
	//
	// Note: it is terminated by '\n'
	s := strings.Split(v, "| COMMANDS: ")

	if len(s) != 2 {
		return nil, fmt.Errorf("error while parsing 'versioncommands'")
	}

	cmds := strings.Split(strings.TrimSuffix(s[1], "\n"), " ")

	return &VersionCommandsResponse{
		Version:  s[0],
		Commands: cmds,
	}, nil
}
