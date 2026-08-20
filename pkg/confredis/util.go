package confredis

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

func HostPort(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty address")
	}
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", err
		}
		if u.Host == "" {
			return "", errors.New("missing host")
		}
		return u.Host, nil
	}
	if _, _, err := net.SplitHostPort(raw); err != nil {
		return "", err
	}
	return raw, nil
}
