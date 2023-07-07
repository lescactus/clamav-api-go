package clamav

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
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
	InStream(ctx context.Context, r io.Reader, size int64) ([]byte, error)
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

// InStream will attempt to connect to Clamd, send the command over the network ("INSTREAM")
// and stream the given io.Reader to let Clamd scan it.
//
// The stream is sent to Clamd in chunks, after INSTREAM, on the same socket on which the command was sent.
//
// It will read the response and return it as a byte slice as well as any error
// encountered.
//
// See https://linux.die.net/man/8/clamd for a detailed explanation of the INSTREAM command.
func (c *ClamavClient) InStream(ctx context.Context, r io.Reader, size int64) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, fmt.Errorf("error while dialing %s/%s: %w", c.network, c.address, err)
	}

	// The format of the chunk is: '<length><data>' where <length> is the size of the following data in bytes
	// expressed as a 4 byte unsigned integer in network byte order and <data> is the actual chunk.
	// Streaming is terminated by sending a zero-length chunk.

	reader := bufio.NewReaderSize(r, 2048)
	writer := bufio.NewWriter(conn)

	// Start scan command.
	_, err = writer.Write(CmdInstream)
	if err != nil {
		return nil, fmt.Errorf("error while writing command to %s/%s: %w", c.network, c.address, err)
	}
	writer.Flush()

	// The size (refered previously as '<length>') must be a byte[] of length 4 - representing a
	// uint32 in a big-endian format (network byte order, tcp standard).
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(size))
	_, err = writer.Write(b)
	if err != nil {
		return nil, fmt.Errorf("error while writing data length to %s/%s: %w", c.network, c.address, err)
	}
	writer.Flush()

	// Streaming the data
	_, err = reader.WriteTo(writer)
	if err != nil {
		resp, e := c.readResponse(conn)
		if e != nil {
			return nil, fmt.Errorf("error while streaming content to %s/%s: %w", c.network, c.address, err)
		}
		return resp, fmt.Errorf("error while streaming content to %s/%s: %w", c.network, c.address, err)
	}

	// Sending 4 bytes to signal the end of the transfer.
	_, err = writer.Write([]byte{'\000', '\000', '\000', '\000'})
	if err != nil {
		return nil, fmt.Errorf("error while writing end of transfer signal to %s/%s: %w", c.network, c.address, err)
	}
	writer.Flush()

	resp, err := c.readResponse(conn)
	if err != nil {
		return nil, err
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}

	return resp, nil
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
	if bytes.EqualFold(msg, RespErrScanFileSizeLimitExceeded) {
		return ErrScanFileSizeLimitExceeded
	}

	if bytes.HasPrefix(msg, []byte("stream: ")) && bytes.HasSuffix(msg, []byte("FOUND")) {
		return ErrVirusFound
	}

	if bytes.Equal(msg, RespErrUnknownCommand) {
		return ErrUnknownCommand
	}

	return nil
}
