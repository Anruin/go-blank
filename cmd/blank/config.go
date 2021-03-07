package main

import (
	"context"
	"github.com/anruin/go-blank/pkg/monitoring"
	"github.com/anruin/go-blank/pkg/names"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

// Service configuration.
type Config struct {
	Monitoring monitoring.Config `mapstructure:"monitoring"`
}

func (c *Config) Initialize(ctx context.Context) (context.Context, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/blank/")
	viper.AddConfigPath("$HOME/.blank")
	viper.AddConfigPath(".")

	c.SetupDefaults()

	// Try to read configuration from the file.
	err := viper.ReadInConfig()

	if err != nil {
		switch err := err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Warnf("configuration file not found: %v", err)
			break
		default:
			log.Fatalf("failed to load configuration: %v", err)
			return nil, err
		}
	}

	if fileUsed := viper.ConfigFileUsed(); fileUsed != "" {
		log.Infof("using configuration file: %s", viper.ConfigFileUsed())
	}

	// Bind environment variables.
	c.SetupEnvironmentVariables()

	err = viper.Unmarshal(&c)
	if err != nil {
		fmt.Printf("failed to decode configuration: %v", err)
	}

	ctx = context.WithValue(ctx, names.CtxConfig, c)

	return ctx, nil
}

func (c *Config) SetupDefaults() {
	// Default monitoring configuration.
	viper.SetDefault(monitoring.CfgHost, "")
	viper.SetDefault(monitoring.CfgPort, "8080")
}

func (c *Config) SetupEnvironmentVariables() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
}
