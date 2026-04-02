package utils

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func ValidateHost(target string) error {
	host := strings.TrimSpace(target)
	if host == "" {
		return fmt.Errorf("target cannot be empty")
	}

	if strings.Contains(host, "://") {
		u, err := url.Parse(host)
		if err != nil {
			return err
		}
		host = u.Hostname()
	}

	if net.ParseIP(host) == nil {
		if len(host) < 3 || !strings.Contains(host, ".") {
			return fmt.Errorf("invalid host or domain: %s", target)
		}
	}

	return nil
}

func NormalizeURL(target string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	return "https://" + target
}

func HostFromTarget(target string) string {
	target = strings.TrimSpace(target)
	if strings.Contains(target, "://") {
		u, err := url.Parse(target)
		if err == nil {
			return u.Hostname()
		}
	}
	return target
}
