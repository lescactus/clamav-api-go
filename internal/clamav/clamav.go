package clamav

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

type Clamaver interface {
	Ping(ctx context.Context) ([]byte, error)
	Version(ctx context.Context) ([]byte, error)
	Reload(ctx context.Context) error
	Stats(ctx context.Context) ([]byte, error)
	VersionCommands(ctx context.Context) ([]byte, error)
	Shutdown(ctx context.Context) error
}

type ClamavClient struct {
	dialer  net.Dialer
	address string
	network string
}

var _ Clamaver = (*ClamavClient)(nil)

func NewClamavClient(addr string, netw string, timeout time.Duration, keepalive time.Duration) *ClamavClient {
	return &ClamavClient{
		dialer: net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepalive,
		},
		address: addr,
		network: netw,
	}
}

func (c *ClamavClient) Ping(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := c.SendCommand(conn, CmdPing)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

func (c *ClamavClient) Version(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := c.SendCommand(conn, CmdVersionBytes)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

func (c *ClamavClient) Reload(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := c.SendCommand(conn, CmdReload)
	if err != nil {
		return fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return fmt.Errorf("error from clamav: %w", err)
	}

	if !bytes.Equal(resp, RespReload) {
		return fmt.Errorf("error from clamav: %w. Expected %s but got %s", ErrUnexpectedResponse, RespReload, resp)
	}
	return nil
}

func (c *ClamavClient) Stats(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := c.SendCommand(conn, CmdStats)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

func (c *ClamavClient) VersionCommands(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := c.SendCommand(conn, CmdVersionCommands)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

func (c *ClamavClient) Shutdown(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = c.SendCommand(conn, CmdVersionCommands)
	if err != nil {
		return fmt.Errorf("error while sending command: %w", err)
	}
	return nil
}

// SendCommand will attempt send the given command to Clamd
// over the network.
// It will read the response and return it as a byte slice as well as any error
// encountered.
//
// See https://linux.die.net/man/8/clamd for a list of supported commands.
func (c *ClamavClient) SendCommand(conn net.Conn, cmd []byte) ([]byte, error) {
	writer := bufio.NewWriter(conn)

	_, err := writer.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("error while writing command to %s/%s: %w", c.network, c.address, err)
	}
	writer.Flush()

	resp, err := c.readResponse(conn)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// readResponse will read from the given io.Reader until a null character is found
// and returns the read bytes before the null character or any error encountered.
func (c *ClamavClient) readResponse(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)

	resp, err := reader.ReadBytes('\000')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error while reading response from %s/%s: %w", c.network, c.address, err)
	}

	// Clamd terminate the response with a NULL character (\000)
	// which can safely be trimed
	return bytes.TrimSuffix(resp, []byte("\000")), nil
}

// parseResponse will attempt to parse the Clamav response to the command
// and determine whether or not Clamav answered with an error.
// See clamav/errors.go for a list of known errors.
func (c *ClamavClient) parseResponse(msg []byte) error {
	if bytes.Equal(msg, RespErrScanFileSizeLimitExceeded) {
		return ErrScanFileSizeLimitExceeded
	}

	if bytes.Contains(msg, []byte("FOUND")) {
		return ErrVirusFound
	}

	if bytes.Equal(msg, RespErrUnknownCommand) {
		return ErrUnknownCommand
	}

	return nil
}
