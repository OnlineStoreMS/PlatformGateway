package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	JWT       JWTConfig
	Upstreams UpstreamsConfig
	CORS      CORSConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ValidatePIM bool   `mapstructure:"validate_pim"`
}

type UpstreamsConfig struct {
	UserCore    string `mapstructure:"usercore"`
	ProductCore string `mapstructure:"productcore"`
}

type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8088
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-me-in-production-use-long-random-string"
	}
	if cfg.Upstreams.UserCore == "" {
		cfg.Upstreams.UserCore = "http://127.0.0.1:8091"
	}
	if cfg.Upstreams.ProductCore == "" {
		cfg.Upstreams.ProductCore = "http://127.0.0.1:8090"
	}
	return &cfg, nil
}

func (c CORSConfig) Allows(origin string) bool {
	if origin == "" {
		return true
	}
	for _, o := range c.AllowOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

func StripTrailingSlash(s string) string {
	return strings.TrimRight(s, "/")
}
