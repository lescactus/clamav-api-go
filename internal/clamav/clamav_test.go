package clamav

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
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
	handlerPing    handlerType = "ping"
	handlerVersion handlerType = "version"
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
	switch handler {
	case handlerPing:
		go s.Serve(handlerPing)
	case handlerVersion:
		go s.Serve(handlerVersion)
	}

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

			go func(conn net.Conn) {
				switch handler {
				case handlerPing:
					s.handlerPing(conn)
					s.wg.Done()
				case handlerVersion:
					s.handlerVersion(conn)
					s.wg.Done()
				default:
					s.handlerPing(conn)
					s.wg.Done()
				}
			}(conn)
		}
	}
}

func (s *ClamdMockTCPServer) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *ClamdMockTCPServer) readFromConnection(conn net.Conn) ([]byte, int) {
	buf := make([]byte, 2048)
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
