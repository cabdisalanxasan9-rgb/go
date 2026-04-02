package subdomain

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"defensekit/internal/worker"
)

type Result struct {
	Subdomain string `json:"subdomain"`
	IP        string `json:"ip,omitempty"`
	Resolved  bool   `json:"resolved"`
}

var defaults = []string{"www", "api", "mail", "dev", "test", "staging", "portal", "cdn", "blog", "vpn"}

func Find(ctx context.Context, domain, wordlistPath string, threads, rate int, timeout time.Duration) []Result {
	words := make([]string, 0, len(defaults)+1000)
	words = append(words, defaults...)

	if strings.TrimSpace(wordlistPath) != "" {
		f, err := os.Open(wordlistPath)
		if err == nil {
			defer f.Close()
			s := bufio.NewScanner(f)
			for s.Scan() {
				w := strings.TrimSpace(s.Text())
				if w != "" {
					words = append(words, w)
				}
			}
		}
	}

	job := func(jobCtx context.Context, word string) Result {
		sub := fmt.Sprintf("%s.%s", word, domain)
		res := Result{Subdomain: sub}
		ips, err := net.DefaultResolver.LookupHost(jobCtx, sub)
		if err != nil || len(ips) == 0 {
			return res
		}
		res.Resolved = true
		res.IP = ips[0]
		return res
	}

	all := worker.Run(ctx, words, threads, rate, timeout, job)
	out := make([]Result, 0, len(all))
	for _, r := range all {
		if r.Resolved {
			out = append(out, r)
		}
	}
	return out
}
