package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppSetDefaults(t *testing.T) {
	app := App{}

	app.setDefaults()

	assert.Equal(t, defaultServerAddr, app.ServerAddr)
	assert.Equal(t, defaultServerReadTimeout, app.ServerReadTimeout)
	assert.Equal(t, defaultServerReadHeaderTimeout, app.ServerReadHeaderTimeout)
	assert.Equal(t, defaultServerWriteTimeout, app.ServerWriteTimeout)
	assert.Equal(t, defaultServerMaxRequestSize, app.ServerMaxRequestSize)

	assert.Equal(t, defaultLoggerLogLevel, app.LoggerLogLevel)
	assert.Equal(t, defaultLoggerDurationFieldUnit, app.LoggerDurationFieldUnit)
	assert.Equal(t, defaultLoggerFormat, app.LoggerFormat)

	assert.Equal(t, defaultClamavAddr, app.ClamavAddr)
	assert.Equal(t, defaultClamavNetwork, app.ClamavNetwork)
	assert.Equal(t, defaultClamavTimeout, app.ClamavTimeout)
	assert.Equal(t, defaultClamavKeepAlive, app.ClamavKeepAlive)
}
