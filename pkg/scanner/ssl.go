package scanner

import (
	"crypto/tls"
	"fmt"
	"time"
)

func ReadSSLCert(host string, port int, timeout time.Duration) *SSLResult {
	result := &SSLResult{Host: host, Port: port}

	address := fmt.Sprintf("%s:%d", host, port)
	dialer := &tls.Dialer{
		Config: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12},
	}

	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer conn.Close()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		result.Error = "failed to cast connection to TLS"
		return result
	}

	_ = tlsConn.SetDeadline(time.Now().Add(timeout))
	if err := tlsConn.Handshake(); err != nil {
		result.Error = err.Error()
		return result
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		result.Error = "no certificate presented"
		return result
	}

	cert := state.PeerCertificates[0]
	now := time.Now()

	result.Subject = cert.Subject.String()
	result.Issuer = cert.Issuer.String()
	result.NotBefore = cert.NotBefore
	result.NotAfter = cert.NotAfter
	result.SerialNumber = cert.SerialNumber.String()
	result.ValidNow = now.After(cert.NotBefore) && now.Before(cert.NotAfter)

	return result
}
