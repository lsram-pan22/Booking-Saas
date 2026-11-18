package config

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"time"
)

func optionalList(value string) []string {
	result := []string{}
	for _, v := range strings.Split(value, ",") {
		if v = strings.TrimSpace(v); v != "" {
			result = append(result, v)
		}
	}
	return result
}
func (c *Config) ValidateSecurity() error {
	secret := strings.TrimSpace(c.JWT.Secret)
	if len(secret) < 32 || strings.Contains(strings.ToLower(secret), "replace-with") || strings.Contains(strings.ToLower(secret), "change-me") {
		return errors.New("JWT_SECRET must contain at least 32 bytes and cannot be a placeholder")
	}
	if c.JWT.ExpiresIn <= 0 || c.JWT.ExpiresIn > 12*time.Hour {
		return errors.New("JWT expiry must be between zero and 12 hours")
	}
	for _, origin := range c.CORSOrigins {
		parsed, err := url.Parse(origin)
		if err != nil || origin != strings.TrimSpace(origin) || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || strings.Contains(origin, "*") {
			return errors.New("CORS_ORIGINS must contain exact HTTP(S) origins without paths or wildcards")
		}
	}
	for _, proxy := range c.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, network, err := net.ParseCIDR(proxy); err != nil || network == nil {
				return errors.New("TRUSTED_PROXIES must contain IP addresses or CIDR networks")
			} else {
				ones, _ := network.Mask.Size()
				if ones == 0 {
					return errors.New("trusting every proxy is not allowed")
				}
			}
		}
	}
	return nil
}
