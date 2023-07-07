package logger

import (
	"os"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	var defaultLogger = zerolog.New(os.Stdout).With().
		Timestamp().
		Logger().Level(zerolog.InfoLevel)

	var defaultLoggerWithErrorLevel = zerolog.New(os.Stdout).With().
		Timestamp().
		Logger().Level(zerolog.ErrorLevel)

	var defaultLoggerWithWarnLevel = zerolog.New(os.Stdout).With().
		Timestamp().
		Logger().Level(zerolog.WarnLevel)

	var defaultLoggerWithConsoleWriter = zerolog.New(os.Stdout).With().
		Timestamp().
		Logger().Level(zerolog.InfoLevel).Output(zerolog.ConsoleWriter{Out: os.Stdout})

	type args struct {
		loglevel          string
		durationFieldUnit string
		format            string
	}
	tests := []struct {
		name string
		args args
		want *zerolog.Logger
	}{
		{
			name: "Empty args",
			args: args{},
			want: &defaultLogger,
		},
		{
			name: "Invalid log level",
			args: args{loglevel: "invalid"},
			want: &defaultLogger,
		},
		{
			name: "Info log level",
			args: args{loglevel: "info"},
			want: &defaultLogger,
		},
		{
			name: "Error log level",
			args: args{loglevel: "error"},
			want: &defaultLoggerWithErrorLevel,
		},
		{
			name: "Warn log level",
			args: args{loglevel: "warn"},
			want: &defaultLoggerWithWarnLevel,
		},
		{
			name: "Format invalid",
			args: args{format: "invalid"},
			want: &defaultLogger,
		},
		{
			name: "Format json",
			args: args{format: "json"},
			want: &defaultLogger,
		},
		{
			name: "Format console",
			args: args{format: "console"},
			want: &defaultLoggerWithConsoleWriter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.loglevel, tt.args.durationFieldUnit, tt.args.format); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
