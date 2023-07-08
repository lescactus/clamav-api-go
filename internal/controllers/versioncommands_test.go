package controllers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHandlerVersionCommands(t *testing.T) {
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
				body:   []byte(`{"clamav_version":"ClamAV 1.0.1/26963/Sat Jul  8 07:27:53 2023","commands":["SCAN","QUIT","RELOAD","PING","CONTSCAN","VERSIONCOMMANDS","VERSION","END","SHUTDOWN","MULTISCAN","FILDES","STATS","IDSESSION","INSTREAM","DETSTATSCLEAR","DETSTATS","ALLMATCHSCAN"]}`),
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
			name: "error is VersionCommandsErrMarshall",
			args: args{
				scenario: ScenarioVersionCommandsErrMarshall,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"error while parsing 'versioncommands'"}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&logger, mockClamav)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.VersionCommands)

			ctx := context.WithValue(context.Background(), MockScenario(""), tt.args.scenario)
			req, err := http.NewRequestWithContext(ctx, "GET", "/rest/v1/versioncommands", nil)
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

func TestVersionCommandsMarshall(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name    string
		args    args
		want    *VersionCommandsResponse
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
			args: args{"ClamAV 1.0.0/26804/Mon Feb  6 08:47:07 2023| COMMANDS: SCAN QUIT RELOAD PING"},
			want: &VersionCommandsResponse{
				Version:  "ClamAV 1.0.0/26804/Mon Feb  6 08:47:07 2023",
				Commands: []string{"SCAN", "QUIT", "RELOAD", "PING"},
			},
			wantErr: false,
		},
		{
			name:    "invalid input - 01",
			args:    args{"ClamAV 1.0.0/26804/Mon Feb  6 08:47:07 2023"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid input - 02",
			args:    args{"invalid input"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := versionCommandsMarshall(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("versionCommandsMarshall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("versionCommandsMarshall() = %v, want %v", got, tt.want)
			}
		})
	}
}
