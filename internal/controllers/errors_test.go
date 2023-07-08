package controllers

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrorResponse(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want *ErrorResponse
	}{
		{
			name: "empty msg",
			args: args{""},
			want: &ErrorResponse{
				Status: "error",
				Msg:    "",
			},
		},
		{
			name: "non empty msg",
			args: args{"foobar"},
			want: &ErrorResponse{
				Status: "error",
				Msg:    "foobar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewErrorResponse(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNetError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "error is nil",
			args: args{err: nil},
			want: false,
		},
		{
			name: "error is not net.Error",
			args: args{err: errors.New("foobar")},
			want: false,
		},
		{
			name: "error is not net.Error",
			args: args{err: &net.OpError{}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNetError(tt.args.err); got != tt.want {
				t.Errorf("isNetError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetErrorResponse(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		status      int
		contentType string
		body        []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "error is nil",
			args: args{nil},
			want: want{200, "", []byte("")},
		},
		{
			name: "error is generic error",
			args: args{errors.New("foobar")},
			want: want{http.StatusInternalServerError, "application/json", []byte(`{"status":"error","msg":"foobar"}`)},
		},
		{
			name: "error is net.Error",
			args: args{&net.OpError{}},
			want: want{http.StatusBadGateway, "application/json", []byte(`{"status":"error","msg":"something wrong happened while communicating with clamav"}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			SetErrorResponse(rr, tt.args.err)

			resp := rr.Result()
			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, tt.want.status, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, body)
		})
	}
}
