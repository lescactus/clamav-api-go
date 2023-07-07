package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

const (
	// Name of the application
	AppName = "clamav-api-go"
)

type Config struct {
	*viper.Viper
}

func New() *Config {
	config := &Config{
		Viper: viper.New(),
	}

	// Set default configurations
	config.setDefaults()

	// Select the .env file
	config.SetConfigName(config.GetString("APP_CONFIG_NAME"))
	config.SetConfigType("dotenv")
	config.AddConfigPath(config.GetString("APP_CONFIG_PATH"))

	// Automatically refresh environment variables
	config.AutomaticEnv()

	// Read configuration
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("failed to read configuration:", err.Error())
			os.Exit(1)
		}
	}

	return config
}

func (config *Config) setDefaults() {
	// Set default App configuration
	config.SetDefault("APP_ADDR", ":8080")
	config.SetDefault("APP_CONFIG_NAME", ".env")
	config.SetDefault("APP_CONFIG_PATH", ".")

	// Server configuration
	config.SetDefault("SERVER_READ_TIMEOUT", 30*time.Second)
	config.SetDefault("SERVER_READ_HEADER_TIMEOUT", 10*time.Second)
	config.SetDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)

	// Logger configuration
	// Available: "trace", "debug", "info", "warn", "error", "fatal", "panic"
	// ref: https://pkg.go.dev/github.com/rs/zerolog@v1.26.1#pkg-variables
	config.SetDefault("LOGGER_LOG_LEVEL", "info")
	config.SetDefault("LOGGER_DURATION_FIELD_UNIT", "ms") // Available: "ms", "millisecond", "s", "second"
	config.SetDefault("LOGGER_FORMAT", "json")            // Available: "json", "console"

	// Clamd configuration
	config.SetDefault("CLAMAV_ADDR", "127.0.0.0:3310")
	config.SetDefault("CLAMAV_NETWORK", "tcp")
	config.SetDefault("CLAMAV_TIMEOUT", 30*time.Second)
	config.SetDefault("CLAMAV_KEEPALIVE", 30*time.Second)
}
