package scanner

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func GrabBanner(conn net.Conn, host string, port int, timeout time.Duration) string {
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if port == 80 || port == 8080 || port == 8000 {
		_, _ = fmt.Fprintf(conn, "HEAD / HTTP/1.0\r\nHost: %s\r\n\r\n", host)
	}

	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}
