package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	// Name of the application
	AppName = "clamav-api-go"

	configNameJSON = "config.json"
	configNameYAML = "config.yaml"
	configNameEnv  = "config.env"
)

var (
	defaultServerAddr              = ":8080"
	defaultServerReadTimeout       = 30 * time.Second
	defaultServerReadHeaderTimeout = 10 * time.Second
	defaultServerWriteTimeout      = 30 * time.Second

	defaultLoggerLogLevel          = "info"
	defaultLoggerDurationFieldUnit = "ms"
	defaultLoggerFormat            = "json"

	defaultClamavAddr      = "127.0.0.1:3310"
	defaultClamavNetwork   = "tcp"
	defaultClamavTimeout   = 30 * time.Second
	defaultClamavKeepAlive = 30 * time.Second
)

type App struct {
	// Address for the server to listen on
	ServerAddr string `json:"server_addr" yaml:"server_addr" mapstructure:"SERVER_ADDR"`

	// Maximum duration for the http server to read the entire request, including the body
	ServerReadTimeout time.Duration `json:"server_read_timeout" yaml:"server_read_timeout" mapstructure:"SERVER_READ_TIMEOUT"`

	// Amount of time the http server allow to read request headers
	ServerReadHeaderTimeout time.Duration `json:"server_read_header_timeout" yaml:"server_read_header_timeout" mapstructure:"SERVER_READ_HEADER_TIMEOUT"`

	// Maximum duration before the http server times out writes of the response
	ServerWriteTimeout time.Duration `json:"server_write_timeout" yaml:"server_write_timeout" mapstructure:"SERVER_WRITE_TIMEOUT"`

	// Logger log level
	// Available: "trace", "debug", "info", "warn", "error", "fatal", "panic"
	// ref: https://pkg.go.dev/github.com/rs/zerolog@v1.26.1#pkg-variables
	LoggerLogLevel string `json:"logger_log_level" yaml:"logger_log_level" mapstructure:"LOGGER_LOG_LEVEL"`

	// Defines the unit for `time.Duration` type fields in the logger
	// Available: "ms", "millisecond", "s", "second"
	LoggerDurationFieldUnit string `json:"logger_duration_field_unit" yaml:"logger_duration_field_unit" mapstructure:"LOGGER_DURATION_FIELD_UNIT"`

	// Format of the logs
	LoggerFormat string `json:"logger_format" yaml:"logger_format" mapstructure:"LOGGER_FORMAT"`

	// Network address of the Clamav server
	ClamavAddr string `json:"clamav_addr" yaml:"clamav_addr" mapstructure:"CLAMAV_ADDR"`

	// Define the named network of the Clamav server
	ClamavNetwork string `json:"clamav_network" yaml:"clamav_network" mapstructure:"CLAMAV_NETWORK"`

	// Maximum amount of time a dial to the Clamav server will wait for a connect to complete
	ClamavTimeout time.Duration `json:"clamav_timeout" yaml:"clamav_timeout" mapstructure:"CLAMAV_TIMEOUT"`

	// Interval between keep-alive probes for an active connection to the Clamav server
	ClamavKeepAlive time.Duration `json:"clamav_keepalive" yaml:"clamav_keepalive" mapstructure:"CLAMAV_KEEPALIVE"`
}

// New will retrieve the runtime configuration from either
// files or environment variables.
//
// Available configuration files are:
//
// * json
//
// * yaml
//
// * dotenv
func New() (*App, error) {
	c := App{}

	// Set default configurations
	c.setDefaults()

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	if _, fileErr := os.Stat(filepath.Join(".", configNameJSON)); fileErr == nil {
		viper.SetConfigType("json")
	} else if _, fileErr := os.Stat(filepath.Join(".", configNameYAML)); fileErr == nil {
		viper.SetConfigType("yaml")
	} else if _, fileErr := os.Stat(filepath.Join(".", configNameEnv)); fileErr == nil {
		viper.SetConfigType("env")
	}

	// When the error is viper.ConfigFileNotFoundError, we try to read from
	// environment variables
	err := viper.ReadInConfig()
	if err != nil {
		switch err.(type) {
		default:
			return nil, fmt.Errorf("error while loading config file: %s", err)
		case viper.ConfigFileNotFoundError:
			readConfigFromEnvVars(c)
		}
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	err = validateConfig(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// validateConfig will make sure the provided configuration is valid
// by looking if the values are present when they are expected to be present
func validateConfig(c *App) error {
	return nil
}

func readConfigFromEnvVars(c App) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvs(c)
}

// ref: https://github.com/spf13/viper/issues/188#issuecomment-399884438
func bindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			bindEnvs(v.Interface(), append(parts, tv)...)
		default:
			viper.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}

func (config *App) setDefaults() {
	config.ServerAddr = defaultServerAddr
	config.ServerReadTimeout = defaultServerReadTimeout
	config.ServerReadHeaderTimeout = defaultServerReadHeaderTimeout
	config.ServerWriteTimeout = defaultServerWriteTimeout

	config.LoggerLogLevel = defaultLoggerLogLevel
	config.LoggerDurationFieldUnit = defaultLoggerDurationFieldUnit
	config.LoggerFormat = defaultLoggerFormat

	config.ClamavAddr = defaultClamavAddr
	config.ClamavNetwork = defaultClamavNetwork
	config.ClamavTimeout = defaultClamavTimeout
	config.ClamavKeepAlive = defaultClamavKeepAlive
}
