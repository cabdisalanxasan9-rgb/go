package scanner

import "time"

type HTTPResult struct {
	URL          string        `json:"url"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

type PortResult struct {
	Port   int    `json:"port"`
	Open   bool   `json:"open"`
	Banner string `json:"banner,omitempty"`
	Error  string `json:"error,omitempty"`
}

type SubdomainResult struct {
	Subdomain string `json:"subdomain"`
	Resolved  bool   `json:"resolved"`
	IP        string `json:"ip,omitempty"`
	Error     string `json:"error,omitempty"`
}

type PasswordStrengthResult struct {
	Length      int      `json:"length"`
	EntropyBits float64  `json:"entropy_bits"`
	Strength    string   `json:"strength"`
	Warnings    []string `json:"warnings,omitempty"`
}

type DNSResult struct {
	Host  string   `json:"host"`
	IPs   []string `json:"ips,omitempty"`
	CNAME string   `json:"cname,omitempty"`
	MX    []string `json:"mx,omitempty"`
	NS    []string `json:"ns,omitempty"`
	TXT   []string `json:"txt,omitempty"`
	Error string   `json:"error,omitempty"`
}

type SSLResult struct {
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	Subject      string    `json:"subject,omitempty"`
	Issuer       string    `json:"issuer,omitempty"`
	NotBefore    time.Time `json:"not_before,omitempty"`
	NotAfter     time.Time `json:"not_after,omitempty"`
	SerialNumber string    `json:"serial_number,omitempty"`
	ValidNow     bool      `json:"valid_now"`
	Error        string    `json:"error,omitempty"`
}

type LatencyResult struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
}

type ToolkitResult struct {
	Warning         string                  `json:"warning"`
	Target          string                  `json:"target,omitempty"`
	HTTP            *HTTPResult             `json:"http,omitempty"`
	Subdomains      []SubdomainResult       `json:"subdomains,omitempty"`
	Ports           []PortResult            `json:"ports,omitempty"`
	Password        *PasswordStrengthResult `json:"password,omitempty"`
	DNS             *DNSResult              `json:"dns,omitempty"`
	SSL             *SSLResult              `json:"ssl,omitempty"`
	Latency         *LatencyResult          `json:"latency,omitempty"`
	GeneratedAtUnix int64                   `json:"generated_at_unix"`
}
