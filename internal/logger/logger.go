package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(loglevel, durationFieldUnit, format string) *zerolog.Logger {
	// Parse loglevel to a zerolog.Level
	// Default to InfoLevel
	level, err := zerolog.ParseLevel(loglevel)
	if err != nil || loglevel == "" {
		level = zerolog.InfoLevel
	}

	// Set the unit for the time.Duration fields
	switch durationFieldUnit {
	case "ms", "millisecond":
		zerolog.DurationFieldUnit = time.Millisecond
	case "s", "second":
		zerolog.DurationFieldUnit = time.Second
	default:
		zerolog.DurationFieldUnit = time.Millisecond
	}

	// Create logger
	l := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger().Level(level)

	// Set the logger to a ConsoleWriter if needed
	switch format {
	case "console":
		l = l.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	case "json":
		break
	default:
		break
	}

	return &l
}
