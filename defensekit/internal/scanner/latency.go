package scanner

import (
	"fmt"
	"net"
	"time"
)

type LatencyResult struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
}

func Latency(host string, port int, timeout time.Duration) *LatencyResult {
	res := &LatencyResult{Host: host, Port: port}
	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer conn.Close()

	res.Latency = time.Since(start)
	return res
}
