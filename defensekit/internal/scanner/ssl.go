package scanner

import (
	"crypto/tls"
	"fmt"
	"time"
)

type SSLResult struct {
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Subject   string    `json:"subject,omitempty"`
	Issuer    string    `json:"issuer,omitempty"`
	NotBefore time.Time `json:"not_before,omitempty"`
	NotAfter  time.Time `json:"not_after,omitempty"`
	ValidNow  bool      `json:"valid_now"`
	Error     string    `json:"error,omitempty"`
}

func SSLInfo(host string, port int, timeout time.Duration) *SSLResult {
	res := &SSLResult{Host: host, Port: port}
	address := fmt.Sprintf("%s:%d", host, port)

	dialer := &tls.Dialer{Config: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer conn.Close()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		res.Error = "tls cast failed"
		return res
	}
	_ = tlsConn.SetDeadline(time.Now().Add(timeout))
	if err := tlsConn.Handshake(); err != nil {
		res.Error = err.Error()
		return res
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		res.Error = "no peer certificate"
		return res
	}

	cert := state.PeerCertificates[0]
	res.Subject = cert.Subject.String()
	res.Issuer = cert.Issuer.String()
	res.NotBefore = cert.NotBefore
	res.NotAfter = cert.NotAfter
	now := time.Now()
	res.ValidNow = now.After(cert.NotBefore) && now.Before(cert.NotAfter)
	return res
}
