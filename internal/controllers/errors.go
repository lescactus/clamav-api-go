package controllers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
)

// ErrorResponse represents the json response
// for http errors
type ErrorResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func NewErrorResponse(msg string) *ErrorResponse {
	return &ErrorResponse{
		Status: "error",
		Msg:    msg,
	}
}

// SetErrorResponse will attempt to parse the given error
// and make a response using the ResponseWriter according to the
// type of the error.
func SetErrorResponse(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var errResp *ErrorResponse

	if isNetError(err) {
		errResp = NewErrorResponse("something wrong happened while communicating with clamav")
		w.WriteHeader(http.StatusBadGateway)
	} else {
		errResp = NewErrorResponse(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	resp, _ := json.Marshal(errResp)
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(resp)
}

// isNetError returns true if the error
// is a net.Error
func isNetError(err error) bool {
	var e net.Error
	return errors.As(err, &e)
}
