package config

import (
	"strings"
	"testing"
	"time"
)

func TestSecurityConfigurationFailsClosed(t *testing.T) {
	valid := func() *Config {
		return &Config{JWT: JWTConfig{Secret: strings.Repeat("a", 40), ExpiresIn: time.Hour}, CORSOrigins: []string{"http://localhost:5173"}, TrustedProxies: []string{"127.0.0.1"}}
	}
	if err := valid().ValidateSecurity(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*Config)
	}{
		{"short secret", func(c *Config) { c.JWT.Secret = "short" }},
		{"placeholder secret", func(c *Config) { c.JWT.Secret = "replace-with-a-long-random-secret" }},
		{"long expiry", func(c *Config) { c.JWT.ExpiresIn = 24 * time.Hour }},
		{"wildcard CORS", func(c *Config) { c.CORSOrigins = []string{"*"} }},
		{"origin path", func(c *Config) { c.CORSOrigins = []string{"https://example.test/path"} }},
		{"origin credentials", func(c *Config) { c.CORSOrigins = []string{"https://user@example.test"} }},
		{"all proxies", func(c *Config) { c.TrustedProxies = []string{"0.0.0.0/0"} }},
		{"invalid proxy", func(c *Config) { c.TrustedProxies = []string{"example.test"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid()
			tc.change(cfg)
			if cfg.ValidateSecurity() == nil {
				t.Fatal("unsafe configuration accepted")
			}
		})
	}
}
