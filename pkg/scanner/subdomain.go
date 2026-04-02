package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"dkk/pkg/pool"
)

var defaultWordlist = []string{
	"www", "mail", "api", "dev", "staging", "test", "admin", "portal", "vpn", "cdn", "blog",
}

func FindSubdomains(ctx context.Context, domain string, externalWordlist string, workers, rate int, timeout time.Duration) []SubdomainResult {
	words := make([]string, 0, len(defaultWordlist)+1000)
	words = append(words, defaultWordlist...)

	if externalWordlist != "" {
		file, err := os.Open(externalWordlist)
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				w := strings.TrimSpace(scanner.Text())
				if w != "" {
					words = append(words, w)
				}
			}
		}
	}

	job := func(jobCtx context.Context, word string) SubdomainResult {
		sub := fmt.Sprintf("%s.%s", word, domain)
		result := SubdomainResult{Subdomain: sub}

		resolver := net.Resolver{}
		ips, err := resolver.LookupHost(jobCtx, sub)
		if err != nil {
			result.Error = err.Error()
			return result
		}

		result.Resolved = true
		if len(ips) > 0 {
			result.IP = ips[0]
		}
		return result
	}

	all := pool.Run(ctx, words, workers, rate, timeout, job)
	found := make([]SubdomainResult, 0, len(all))
	for _, r := range all {
		if r.Resolved {
			found = append(found, r)
		}
	}

	return found
}
