package controllers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHandlerInStream(t *testing.T) {
	logger := zerolog.New(io.Discard)
	mockClamav := &MockClamav{}

	type args struct {
		scenario    MockScenario
		headers     map[string]string
		filename    string
		filecontent string
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
			name: "no error - empty file",
			args: args{
				scenario:    ScenarioNoError,
				filename:    "test.txt",
				filecontent: "",
			},
			want: want{
				status: http.StatusOK,
				body:   []byte(`{"status":"noerror","msg":"stream: OK","signature":"","virus_found":false}`),
			},
		},
		{
			name: "no error",
			args: args{
				scenario:    ScenarioNoError,
				filename:    "test.txt",
				filecontent: "foobar",
			},
			want: want{
				status: http.StatusOK,
				body:   []byte(`{"status":"noerror","msg":"stream: OK","signature":"","virus_found":false}`),
			},
		},
		{
			name: "error is net error",
			args: args{
				scenario:    ScenarioNetError,
				filename:    "test.txt",
				filecontent: "",
			},
			want: want{
				status: http.StatusBadGateway,
				body:   []byte(`{"status":"error","msg":"something wrong happened while communicating with clamav"}`),
			},
		},
		{
			name: "error is ErrUnknownCommand",
			args: args{
				scenario:    ScenarioErrUnknownCommand,
				filename:    "test.txt",
				filecontent: "",
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"unknown command sent to clamav"}`),
			},
		},
		{
			name: "error is ErrUnknownResponse",
			args: args{
				scenario:    ScenarioErrUnknownResponse,
				filename:    "test.txt",
				filecontent: "",
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"unknown response from clamav"}`),
			},
		},
		{
			name: "error is ErrUnexpectedResponse",
			args: args{
				scenario:    ScenarioErrUnexpectedResponse,
				filename:    "test.txt",
				filecontent: "",
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"unexpected response from clamav"}`),
			},
		},
		{
			name: "error is ErrScanFileSizeLimitExceeded",
			args: args{
				scenario:    ScenarioErrScanFileSizeLimitExceeded,
				filename:    "test.txt",
				filecontent: "",
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","msg":"clamav: size limit exceeded"}`),
			},
		},
		{
			name: "error is ErrMissingFile",
			args: args{
				scenario:    ScenarioNoError,
				filename:    "",
				filecontent: "",
			},
			want: want{
				status: http.StatusBadRequest,
				body:   []byte(`{"status":"error","msg":"bad request: failed to parse file: http: no such file"}`),
			},
		},
		{
			name: "error is ErrVirusFound",
			args: args{
				scenario:    ScenarioErrVirusFound,
				filename:    "eicar.txt",
				filecontent: "",
			},
			want: want{
				status: http.StatusOK,
				body:   []byte(`{"status":"error","msg":"file contains potential virus","signature":"","virus_found":true}`),
			},
		},
		{
			name: "content type is set to text/plain",
			args: args{
				scenario:    ScenarioNoError,
				filename:    "test.txt",
				filecontent: "",
				headers:     map[string]string{"Content-Type": "text/plain"},
			},
			want: want{
				status: http.StatusBadRequest,
				body:   []byte(`{"status":"error","msg":"bad request: failed to parse file: request Content-Type isn't multipart/form-data"}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&logger, mockClamav)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.InStream)

			r := strings.NewReader(tt.args.filecontent)
			b := &bytes.Buffer{}
			writer := multipart.NewWriter(b)
			part, _ := writer.CreateFormFile("file", tt.args.filename)
			io.Copy(part, r)
			writer.Close()

			ctx := context.WithValue(context.Background(), MockScenario(""), tt.args.scenario)
			req, err := http.NewRequestWithContext(ctx, "POST", "/rest/v1/scan", b)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", writer.FormDataContentType())
			for k, v := range tt.args.headers {
				req.Header.Set(k, v)
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

func TestHandlerParseSignature(t *testing.T) {
	type fields struct {
		Clamav clamav.Clamaver
		Logger *zerolog.Logger
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "Empty string",
			fields: fields{},
			args:   args{msg: ""},
			want:   "",
		},
		{
			name:   "stream: OK",
			fields: fields{},
			args:   args{msg: "stream: OK"},
			want:   "OK",
		},
		{
			name:   "OK",
			fields: fields{},
			args:   args{msg: "OK"},
			want:   "OK",
		},
		{
			name:   "stream: Eicar-Signature FOUND",
			fields: fields{},
			args:   args{msg: "stream: Eicar-Signature FOUND"},
			want:   "Eicar-Signature",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Clamav: tt.fields.Clamav,
				Logger: tt.fields.Logger,
			}
			if got := h.parseSignature(tt.args.msg); got != tt.want {
				t.Errorf("Handler.parseSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
