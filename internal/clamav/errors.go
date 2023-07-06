package clamav

import "errors"

var (
	ErrUnknownCommand            = errors.New("unknown command")
	ErrUnknownResponse           = errors.New("unknown response from clamav")
	ErrUnexpectedResponse        = errors.New("unexpected response from clamav")
	ErrScanFileSizeLimitExceeded = errors.New("size limit exceeded")
	ErrVirusFound                = errors.New("file contains potential virus")
)
