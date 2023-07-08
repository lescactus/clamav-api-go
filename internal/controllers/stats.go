package controllers

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/hlog"
)

// StatsResponse represents the json response of a /stats endpoint.
// It represents the statistics about the scan queue,
// contents of scan queue, and memory usage.
type StatsResponse struct {
	Pools    int    `json:"pools"`
	State    string `json:"state"`
	Threads  string `json:"threads"`
	Queue    string `json:"queue"`
	Memstats string `json:"memstats"`
}

var ErrParsingStats = errors.New("error while parsing 'stats'")

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	stats, err := h.Clamav.Stats(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msgf("error while sending stats command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("stats command sent successfully")

	s, err := statsMarshall(string(stats))
	if err != nil {
		h.Logger.Error().Str("req_id", req_id.String()).Msgf("error while marshalling stats: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", req_id.String()).Msg("stats marshalled successfully")

	resp, err := json.Marshal(&s)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// statsMarshall will marshall the string s
// into a *StatsResponse.
//
// Example of output for a 'STATS' command:
//
// POOLS: 1
//
// STATE: VALID PRIMARY
// THREADS: live 1  idle 0 max 10 idle-timeout 30
// QUEUE: 0 items
//
//	STATS 0.000111
//
// MEMSTATS: heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 713.137M pools_total 713.226M
// END
//
// It returns any error encountered.
func statsMarshall(s string) (*StatsResponse, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("error: empty string")
	}

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)

	var pools int
	var state string
	var threads string
	var queue string
	var memstats string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "POOLS: ") {
			var err error
			s := strings.SplitAfter(line, ": ")
			if len(s) != 2 {
				return nil, ErrParsingStats
			}
			pools, err = strconv.Atoi(s[1])
			if err != nil {
				return nil, ErrParsingStats
			}
		}

		if strings.HasPrefix(line, "STATE: ") {
			s := strings.SplitAfter(line, ": ")
			if len(s) != 2 {
				return nil, ErrParsingStats
			}
			state = s[1]
		}

		if strings.HasPrefix(line, "THREADS: ") {
			s := strings.SplitAfter(line, ": ")
			if len(s) != 2 {
				return nil, ErrParsingStats
			}
			threads = s[1]
		}

		// "QUEUE" is a multi-line string until the "MEMSTATS" section
		if strings.HasPrefix(line, "QUEUE: ") {
			re := regexp.MustCompile(`(?ms)QUEUE: (.*?)\sMEMSTATS`)
			matches := re.FindStringSubmatch(s)

			if len(matches) != 2 {
				return nil, ErrParsingStats
			}
			queue = strings.TrimSuffix(matches[1], "\n")
		}

		if strings.HasPrefix(line, "MEMSTATS: ") {
			s := strings.SplitAfter(line, ": ")
			if len(s) != 2 {
				return nil, ErrParsingStats
			}
			memstats = s[1]
		}
	}

	return &StatsResponse{
		Pools:    pools,
		State:    state,
		Threads:  threads,
		Queue:    queue,
		Memstats: memstats,
	}, nil
}
