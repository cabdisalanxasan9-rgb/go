package scanner

import (
	"context"
	"net"
	"strconv"
	"time"

	"defensekit/internal/worker"
)

type PortResult struct {
	Port   int    `json:"port"`
	Open   bool   `json:"open"`
	Banner string `json:"banner,omitempty"`
}

func PortScan(ctx context.Context, host string, start, end, threads, rate int, timeout time.Duration) []PortResult {
	ports := make([]int, 0, end-start+1)
	for p := start; p <= end; p++ {
		ports = append(ports, p)
	}

	job := func(_ context.Context, p int) PortResult {
		address := net.JoinHostPort(host, strconv.Itoa(p))
		res := PortResult{Port: p}
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return res
		}
		defer conn.Close()

		res.Open = true
		res.Banner = GrabBanner(conn, host, p, timeout)
		return res
	}

	all := worker.Run(ctx, ports, threads, rate, timeout, job)
	open := make([]PortResult, 0, len(all))
	for _, r := range all {
		if r.Open {
			open = append(open, r)
		}
	}
	return open
}
