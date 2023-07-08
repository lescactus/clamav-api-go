package controllers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHandlerPing(t *testing.T) {
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
				body:   []byte(`{"ping":"PONG"}`),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&logger, mockClamav)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.Ping)

			ctx := context.WithValue(context.Background(), MockScenario(""), tt.args.scenario)
			req, err := http.NewRequestWithContext(ctx, "GET", "/rest/v1/ping ", nil)
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
