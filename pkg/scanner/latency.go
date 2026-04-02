package scanner

import (
	"net"
	"strconv"
	"time"
)

func CheckLatency(host string, port int, timeout time.Duration) *LatencyResult {
	result := &LatencyResult{Host: host, Port: port}
	address := net.JoinHostPort(host, strconv.Itoa(port))

	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer conn.Close()

	result.Latency = time.Since(start)
	return result
}
