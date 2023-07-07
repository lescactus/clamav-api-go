package clamav

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	// specifying the port ':0' will leave the kernel chosing a
	// random port to assign
	listen  = "127.0.0.1:0"
	network = "tcp"
)

// handlerType represents the type of handler
// the mock tcp server need to implement
type handlerType string

var (
	handlerPing                handlerType = "ping"
	handlerVersion             handlerType = "version"
	handlerReload              handlerType = "reload"
	handlerStats               handlerType = "stats"
	handlerVersionCommands     handlerType = "versioncommands"
	handlerShutdown            handlerType = "shutdown"
	handlerInStreamGoodFile    handlerType = "instreamgoodfile"
	handlerInStreamBadFile     handlerType = "instreamgbadfile"
	handlerInStreamTooLongFile handlerType = "instreamtoolongfile"
)

// ClamdMockTCPServer is a tcp server
// mocking the clamd daemon by implementing basic
// commands handlers such as Ping, Version, Stats, etc...
type ClamdMockTCPServer struct {
	listener net.Listener
	quit     chan struct{}
	ready    chan bool
	wg       sync.WaitGroup
}

// Mostly taken from https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
func NewServer(netw, addr string, handler handlerType) *ClamdMockTCPServer {
	s := &ClamdMockTCPServer{
		quit:  make(chan struct{}),
		ready: make(chan bool, 1),
	}
	l, err := net.Listen(netw, addr)
	if err != nil {
		log.Fatal(err)
	}
	s.listener = l
	s.wg.Add(1)

	go s.Serve(handler)

	s.ready <- true
	return s
}

func (s *ClamdMockTCPServer) Serve(handler handlerType) {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Println("accept error", err)
			}
		} else {
			s.wg.Add(1)

			go func(handler handlerType, conn net.Conn) {
				switch handler {
				case handlerPing:
					s.handlerPing(conn)
					s.wg.Done()
				case handlerVersion:
					s.handlerVersion(conn)
					s.wg.Done()
				case handlerReload:
					s.handlerReload(conn)
					s.wg.Done()
				case handlerStats:
					s.handlerStats(conn)
					s.wg.Done()
				case handlerVersionCommands:
					s.handlerVersionCommands(conn)
					s.wg.Done()
				case handlerShutdown:
					s.handlerShutdown(conn)
					s.wg.Done()
				case handlerInStreamGoodFile:
					s.handlerInStreamGoodFile(conn)
					s.wg.Done()
				case handlerInStreamBadFile:
					s.handlerInStreamBadFile(conn)
					s.wg.Done()
				case handlerInStreamTooLongFile:
					s.handlerInStreamTooLongFile(conn)
					s.wg.Done()
				default:
					s.handlerPing(conn)
					s.wg.Done()
				}
			}(handler, conn)
		}
	}
}

func (s *ClamdMockTCPServer) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *ClamdMockTCPServer) readFromConnection(conn net.Conn) ([]byte, int) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("read error", err)
			return nil, 0
		}
		if n == 0 {
			return nil, 0
		}
		return buf, n
	}

}

func (s *ClamdMockTCPServer) handlerPing(conn net.Conn) {
	defer conn.Close()

	fmt.Fprint(conn, "PONG\000")

}

func (s *ClamdMockTCPServer) handlerVersion(conn net.Conn) {
	defer conn.Close()

	//log.Printf("received from %v: %s", conn.RemoteAddr(), string(buf[:n]))
	fmt.Fprintf(conn, "ClamAV 1.0.1/26961/%s\000", time.Now().Format("Thu Jul  6 07:29:38 2023"))

}

func (s *ClamdMockTCPServer) handlerReload(conn net.Conn) {
	defer conn.Close()

	fmt.Fprint(conn, "RELOADING\000")
}

// Example of output for a 'STATS' command
var statsResp = `POOLS: 1

STATE: VALID PRIMARY
THREADS: live 1  idle 0 max 10 idle-timeout 30
QUEUE: 0 items
	STATS 0.000038

MEMSTATS: heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 1307.045M pools_total 1307.093M
END`

func (s *ClamdMockTCPServer) handlerStats(conn net.Conn) {
	defer conn.Close()

	s.readFromConnection(conn)
	fmt.Fprint(conn, statsResp)
}

// Example of output for a 'VERSIONCOMMANDS' command
var versionCommandsResp = `ClamAV 1.0.1/26961/Thu Jul  6 07:29:38 2023| COMMANDS: SCAN QUIT RELOAD PING CONTSCAN VERSIONCOMMANDS VERSION END SHUTDOWN MULTISCAN FILDES STATS IDSESSION INSTREAM DETSTATSCLEAR DETSTATS ALLMATCHSCAN`

func (s *ClamdMockTCPServer) handlerVersionCommands(conn net.Conn) {
	defer conn.Close()

	s.readFromConnection(conn)
	fmt.Fprint(conn, versionCommandsResp)
}

func (s *ClamdMockTCPServer) handlerShutdown(conn net.Conn) {
	defer conn.Close()
}

var (
	goodFile    = `foobar`
	badFile     = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`
	tooLongFile = `db94dcc6543c3b1fe3c3ca1948c5ac465998d7ac830535dd54f385142182072152000e398c4554133ca3038467f976d654d23bb3a7480a01bb681d8f6be8ce6az`
)

func (s *ClamdMockTCPServer) handlerInStreamGoodFile(conn net.Conn) {
	defer conn.Close()

	// Clamav expects a stream of chunks. The format of the chunk is: '<length><data>'
	// where <length> is the size of the following data in bytes expressed as a 4 byte unsigned
	// integer in network byte order and <data> is the actual chunk.
	// Streaming is terminated by sending a zero-length chunk.
	//
	// For example, to send 'foobar', the bytes sent would be:
	// [122 73 78 83 84 82 69 65 77 0 0 0 0 6 102 111 111 98 97 114 0 0 0 0]
	//   ^                       ^  ^       ^  ^                 ^  ^     ^
	//   |       zINSTREAM       |  |  len  |  |      foobar     |  | null|
	//   +-----------------------+  +-------+  +-----------------+  +-----+
	buf := make([]byte, 64)
	msg := make([]byte, 0)

	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatalf("error while reading response: %s", err)
		}

		msg = append(msg, buf[:n]...)
		if err == io.EOF || n == 0 || bytes.HasSuffix(msg, []byte{'\000', '\000', '\000', '\000'}) {
			break
		}
	}

	fmt.Fprint(conn, "stream: OK\000")
}

func (s *ClamdMockTCPServer) handlerInStreamBadFile(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 64)
	msg := make([]byte, 0)

	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatalf("error while reading response: %s", err)
		}

		msg = append(msg, buf[:n]...)
		if err == io.EOF || n == 0 || bytes.HasSuffix(msg, []byte{'\000', '\000', '\000', '\000'}) {
			break
		}
	}

	fmt.Fprint(conn, "stream: Win.Test.EICAR_HDB-1 FOUND\000")
}

func (s *ClamdMockTCPServer) handlerInStreamTooLongFile(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 64)
	msg := make([]byte, 0)

	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatalf("error while reading response: %s", err)
		}

		msg = append(msg, buf[:n]...)
		if err == io.EOF || n == 0 || bytes.HasSuffix(msg, []byte{'\000', '\000', '\000', '\000'}) {
			break
		}
	}
	// Get file size. It's 4 bytes after the Command
	fsize := msg[len(CmdInstream) : len(CmdInstream)+4]
	size := binary.BigEndian.Uint32(fsize)

	if size > 128 {
		fmt.Fprint(conn, "INSTREAM size limit exceeded. ERROR\000")
	}
}

func TestNewClamavClient(t *testing.T) {
	type args struct {
		addr      string
		netw      string
		timeout   time.Duration
		keepalive time.Duration
	}
	tests := []struct {
		name string
		args args
		want *ClamavClient
	}{
		{
			name: "empty args",
			args: args{"", "", 0, 0},
			want: &ClamavClient{dialer: net.Dialer{}, address: "", network: ""},
		},
		{
			name: "address set - empty args",
			args: args{"127.0.0.1", "", 0, 0},
			want: &ClamavClient{dialer: net.Dialer{}, address: "127.0.0.1", network: ""},
		},
		{
			name: "network set - empty args",
			args: args{"", "tcp", 0, 0},
			want: &ClamavClient{dialer: net.Dialer{}, address: "", network: "tcp"},
		},
		{
			name: "timeout set to 10s - empty args",
			args: args{"", "", 10 * time.Second, 0},
			want: &ClamavClient{dialer: net.Dialer{Timeout: 10 * time.Second, KeepAlive: 0}},
		},
		{
			name: "keepalive set to 10s - empty args",
			args: args{"", "", 0, 10 * time.Second},
			want: &ClamavClient{dialer: net.Dialer{Timeout: 0, KeepAlive: 10 * time.Second}},
		},
		{
			name: "address set - network set - timeout set - keepalive set",
			args: args{"127.0.0.1", "tcp", 10 * time.Second, 10 * time.Second},
			want: &ClamavClient{dialer: net.Dialer{Timeout: 10 * time.Second, KeepAlive: 10 * time.Second}, address: "127.0.0.1", network: "tcp"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClamavClient(tt.args.addr, tt.args.netw, tt.args.timeout, tt.args.keepalive)
			assert.Equal(t, tt.want.address, c.address)
			assert.Equal(t, tt.want.network, c.network)
			assert.Equal(t, tt.want.dialer.Timeout, c.dialer.Timeout)
			assert.Equal(t, tt.want.dialer.KeepAlive, c.dialer.KeepAlive)
		})
	}
}

func TestClamavClientPing(t *testing.T) {
	// Start mock tcp server on random port and wait for it to be ready
	s := NewServer(network, listen, handlerPing)
	<-s.ready

	c := NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	resp, err := c.Ping(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []byte(RespPing), resp)

	// Stop mock tcp server
	s.Stop()

	// When the server is stopped
	resp, err = c.Ping(context.Background())
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestClamavClientVersion(t *testing.T) {
	// Start mock tcp server on random port and wait for it to be ready
	s := NewServer(network, listen, handlerVersion)
	<-s.ready

	c := NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	// Clamav typical answer to the "version" command might be:
	// "ClamAV 1.0.1/26961/Thu Jul  6 07:29:38 2023"
	versionRespRegex := `ClamAV\ [0-9]+.[0-9]+.[0-9]+\/[0-9]+\/[a-zA-Z]+\ [a-zA-Z]+.*`
	resp, err := c.Version(context.Background())
	assert.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(versionRespRegex), string(resp))

	// Stop mock tcp server
	s.Stop()

	// When the server is stopped
	resp, err = c.Version(context.Background())
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestClamavClientReload(t *testing.T) {
	// Start mock tcp server on random port and wait for it to be ready
	s := NewServer(network, listen, handlerReload)
	<-s.ready

	c := NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	err := c.Reload(context.Background())
	assert.NoError(t, err)

	// Stop mock tcp server
	s.Stop()

	// When the server is stopped
	err = c.Reload(context.Background())
	assert.Error(t, err)
}

func TestClamavClientStats(t *testing.T) {
	// Start mock tcp server on random port and wait for it to be ready
	s := NewServer(network, listen, handlerStats)
	<-s.ready

	c := NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	resp, err := c.Stats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, statsResp, string(resp))

	// Stop mock tcp server
	s.Stop()

	// When the server is stopped
	resp, err = c.Stats(context.Background())
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestClamavClientVersionCommands(t *testing.T) {
	// Start mock tcp server on random port and wait for it to be ready
	s := NewServer(network, listen, handlerVersionCommands)
	<-s.ready

	c := NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	resp, err := c.VersionCommands(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, versionCommandsResp, string(resp))

	// Stop mock tcp server
	s.Stop()

	// When the server is stopped
	resp, err = c.VersionCommands(context.Background())
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestClamavClientVersionShutdown(t *testing.T) {
	// Start mock tcp server on random port and wait for it to be ready
	s := NewServer(network, listen, handlerShutdown)
	<-s.ready

	c := NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	err := c.Shutdown(context.Background())
	assert.NoError(t, err)

	// Stop mock tcp server
	s.Stop()

	// When the server is stopped
	err = c.Shutdown(context.Background())
	assert.Error(t, err)
}

func TestClamavClientInStream(t *testing.T) {
	// Good file
	// Start mock tcp server on random port and wait for it to be ready
	s := NewServer(network, listen, handlerInStreamGoodFile)
	<-s.ready

	c := NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	resp, err := c.InStream(context.Background(), strings.NewReader(goodFile), int64(len(goodFile)))
	assert.EqualValues(t, RespScan, resp)
	assert.NoError(t, err)

	// Stop mock tcp server
	s.Stop()

	// Bad file
	s = NewServer(network, listen, handlerInStreamBadFile)
	<-s.ready

	c = NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	resp, err = c.InStream(context.Background(), strings.NewReader(badFile), int64(len(badFile)))
	assert.True(t, bytes.Contains([]byte("FOUND"), resp))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrVirusFound)

	// Stop mock tcp server
	s.Stop()

	// File is too long
	s = NewServer(network, listen, handlerInStreamTooLongFile)
	<-s.ready

	c = NewClamavClient(s.listener.Addr().String(), s.listener.Addr().Network(),
		time.Second, time.Second)

	_, err = c.InStream(context.Background(), strings.NewReader(tooLongFile), int64(len(tooLongFile)))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrScanFileSizeLimitExceeded)

	// Stop mock tcp server
	s.Stop()

	// When the server is stopped
	resp, err = c.InStream(context.Background(), strings.NewReader(goodFile), int64(len(goodFile)))
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestClamavClientParseResponse(t *testing.T) {
	tests := []struct {
		name    string
		resp    []byte
		wantErr bool
		typeErr error
	}{
		{
			name:    "empty response",
			resp:    []byte(""),
			wantErr: false,
			typeErr: nil,
		},
		{
			name:    "response is foobar",
			resp:    []byte("foobar"),
			wantErr: false,
			typeErr: nil,
		},
		{
			name:    "response is PONG",
			resp:    RespPing,
			wantErr: false,
			typeErr: nil,
		},
		{
			name:    "response is RELOADING",
			resp:    RespReload,
			wantErr: false,
			typeErr: nil,
		},
		{
			name:    "response is stream: OK",
			resp:    RespScan,
			wantErr: false,
			typeErr: nil,
		},
		{
			name:    "response is UNKNOWN COMMAND",
			resp:    RespErrUnknownCommand,
			wantErr: true,
			typeErr: ErrUnknownCommand,
		},
		{
			name:    "response is stream: Eicar FOUND",
			resp:    []byte("stream: Eicar FOUND"),
			wantErr: true,
			typeErr: ErrVirusFound,
		},
		{
			name:    "response is INSTREAM size limit exceeded. ERROR",
			resp:    RespErrScanFileSizeLimitExceeded,
			wantErr: true,
			typeErr: ErrScanFileSizeLimitExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ClamavClient{}

			err := c.parseResponse(tt.resp)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.typeErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
