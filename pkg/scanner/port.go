package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"dkk/pkg/pool"
)

func ScanPorts(ctx context.Context, host string, start, end, workers, rate int, timeout time.Duration) []PortResult {
	ports := make([]int, 0, end-start+1)
	for p := start; p <= end; p++ {
		ports = append(ports, p)
	}

	job := func(_ context.Context, port int) PortResult {
		address := net.JoinHostPort(host, strconv.Itoa(port))
		result := PortResult{Port: port}

		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return result
		}
		defer conn.Close()

		result.Open = true
		result.Banner = grabBanner(conn, host, port, timeout)
		return result
	}

	all := pool.Run(ctx, ports, workers, rate, timeout, job)
	open := make([]PortResult, 0, len(all))
	for _, r := range all {
		if r.Open {
			open = append(open, r)
		}
	}

	return open
}

func grabBanner(conn net.Conn, host string, port int, timeout time.Duration) string {
	_ = conn.SetDeadline(time.Now().Add(timeout))

	if port == 80 || port == 8080 || port == 8000 {
		_, _ = fmt.Fprintf(conn, "HEAD / HTTP/1.0\r\nHost: %s\r\n\r\n", host)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(line)
}
