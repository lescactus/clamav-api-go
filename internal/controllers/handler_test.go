package controllers

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog"
)

func TestNewHandler(t *testing.T) {
	logger := zerolog.Logger{}
	c := clamav.ClamavClient{}
	type args struct {
		logger *zerolog.Logger
		clamav clamav.Clamaver
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "nil args",
			args: args{nil, nil},
			want: &Handler{nil, nil},
		},
		{
			name: "non nil args",
			args: args{&logger, &c},
			want: &Handler{&c, &logger},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.logger, tt.args.clamav); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockClamav struct{}

var _ clamav.Clamaver = (*MockClamav)(nil)

func (m *MockClamav) Ping(ctx context.Context) ([]byte, error) {
	scenario := ctx.Value(MockScenario(""))

	if scenario == ScenarioNoError {
		return []byte("PONG"), nil
	} else {
		return nil, dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) Version(ctx context.Context) ([]byte, error) {
	scenario := ctx.Value(MockScenario(""))

	if scenario == ScenarioNoError {
		return []byte("ClamAV 1.0.1/26961/Thu Jul  6 07:29:38 2023"), nil
	} else {
		return nil, dispatchErrFromScenario(scenario.(MockScenario))
	}
}

func (m *MockClamav) Reload(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClamav) Stats(ctx context.Context) ([]byte, error) {
	panic("not implemented")
}

func (m *MockClamav) VersionCommands(ctx context.Context) ([]byte, error) {
	panic("not implemented")
}

func (m *MockClamav) Shutdown(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClamav) InStream(ctx context.Context, r io.Reader, size int64) ([]byte, error) {
	panic("not implemented")
}

func dispatchErrFromScenario(scenario MockScenario) error {
	switch scenario {
	case ScenarioNetError:
		return &net.OpError{}
	case ScenarioErrUnknownCommand:
		return errors.New("unknown command sent to clamav")
	case ScenarioErrUnknownResponse:
		return errors.New("unknown response from clamav")
	case ScenarioErrUnexpectedResponse:
		return errors.New("unexpected response from clamav")
	case ScenarioErrScanFileSizeLimitExceeded:
		return errors.New("clamav: size limit exceeded")
	default:
		return nil
	}
}

type MockScenario string

var (
	ScenarioNoError                      MockScenario = "noerror"
	ScenarioNetError                     MockScenario = "neterror"
	ScenarioErrUnknownCommand            MockScenario = "unknowncommand"
	ScenarioErrUnknownResponse           MockScenario = "unknownresponse"
	ScenarioErrUnexpectedResponse        MockScenario = "unexpectedresponse"
	ScenarioErrScanFileSizeLimitExceeded MockScenario = "scanfilesizelimitexceeded"
)
