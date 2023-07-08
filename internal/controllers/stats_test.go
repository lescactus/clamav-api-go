package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type fakeValue struct{}

func (f fakeValue) MarshalJSON() ([]byte, error) { return nil, errors.New("") }

func TestHandlerStats(t *testing.T) {
	logger := zerolog.New(io.Discard)
	mockClamav := &MockClamav{}

	type args struct {
		scenario MockScenario
	}
	type want struct {
		status int
		body   []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "no error",
			args: args{
				scenario: ScenarioNoError,
			},
			want: want{
				status: http.StatusOK,
				body:   []byte(`{"pools":1,"state":"VALID PRIMARY","threads":"live 1  idle 0 max 10 idle-timeout 30","queue":"0 items\n\tSTATS 0.000086 ","memstats":"heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 1306.837M pools_total 1306.882M"}`),
			},
		},
		{
			name: "error is net error",
			args: args{
				scenario: ScenarioNetError,
			},
			want: want{
				status: http.StatusBadGateway,
				body:   []byte(`{"status":"error","msg":"something wrong happened while communicating with clamav"}`),
			},
		},
		{
			name: "error is ErrUnknownCommand",
			args: args{
				scenario: ScenarioErrUnknownCommand,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"unknown command sent to clamav"}`),
			},
		},
		{
			name: "error is ErrUnknownResponse",
			args: args{
				scenario: ScenarioErrUnknownResponse,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"unknown response from clamav"}`),
			},
		},
		{
			name: "error is ErrUnexpectedResponse",
			args: args{
				scenario: ScenarioErrUnexpectedResponse,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"unexpected response from clamav"}`),
			},
		},
		{
			name: "error is ErrScanFileSizeLimitExceeded",
			args: args{
				scenario: ScenarioErrScanFileSizeLimitExceeded,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"clamav: size limit exceeded"}`),
			},
		},
		{
			name: "error is ErrParsingStats",
			args: args{
				scenario: ScenarioStatsErrUnmarshall,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(fmt.Sprintf(`{"status":"error","msg":"%s"}`, ErrParsingStats)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&logger, mockClamav)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.Stats)

			ctx := context.WithValue(context.Background(), MockScenario(""), tt.args.scenario)
			req, err := http.NewRequestWithContext(ctx, "GET", "/rest/v1/version ", nil)
			if err != nil {
				t.Fatal(err)
			}

			handler.ServeHTTP(rr, req)

			resp := rr.Result()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.want.status, resp.StatusCode)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, body)
		})
	}
}

func TestStatsMarshall(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *StatsResponse
		wantErr bool
	}{
		{
			name:    "empty input",
			args:    args{""},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid input",
			args: args{
				s: `POOLS: 1

STATE: VALID PRIMARY
THREADS: live 1  idle 0 max 10 idle-timeout 30
QUEUE: 0 items
	STATS 0.000042

MEMSTATS: heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 713.137M pools_total 713.226M
END
				`,
			},
			want: &StatsResponse{
				Pools:    1,
				State:    "VALID PRIMARY",
				Threads:  "live 1  idle 0 max 10 idle-timeout 30",
				Queue:    "0 items\n\tSTATS 0.000042",
				Memstats: "heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 713.137M pools_total 713.226M",
			},
			wantErr: false,
		},
		{
			name: "invalid input - 01",
			args: args{
				s: `POOLS: : 1
END
				`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid input - 02",
			args: args{
				s: `STATE: : 1
END
				`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid input - 03",
			args: args{
				s: `THREADS: : 1
END
				`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid input - 04",
			args: args{
				s: `MEMSTATS: : 1
END
				`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid input - 05",
			args: args{
				s: `QUEUE: : 1
END
				`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid input - 06",
			args: args{
				s: `POOLS: str
END
				`,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := statsMarshall(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("statsMarshall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("statsMarshall() = %v, want %v", got, tt.want)
			}
		})
	}
}
