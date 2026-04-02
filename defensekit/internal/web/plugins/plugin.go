package plugins

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"defensekit/internal/password"
	"defensekit/internal/scanner"
	"defensekit/internal/subdomain"
)

type RunRequest struct {
	Target    string        `json:"target"`
	Threads   int           `json:"threads"`
	Timeout   time.Duration `json:"timeout"`
	Rate      int           `json:"rate"`
	StartPort int           `json:"start_port"`
	EndPort   int           `json:"end_port"`
	Password  string        `json:"password,omitempty"`
}

type Result struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Data      any       `json:"data,omitempty"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	Duration  string    `json:"duration"`
}

type Plugin interface {
	Name() string
	Description() string
	Run(context.Context, RunRequest) (any, error)
}

type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	results map[string]Result
}

func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		results: make(map[string]Result),
	}
}

func (m *Manager) Register(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins[p.Name()] = p
}

func (m *Manager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (m *Manager) Describe() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.plugins))
	for name, p := range m.plugins {
		out[name] = p.Description()
	}
	return out
}

func (m *Manager) RunOne(ctx context.Context, name string, req RunRequest) Result {
	m.mu.RLock()
	p, ok := m.plugins[name]
	m.mu.RUnlock()
	if !ok {
		res := Result{Name: name, Status: "error", Error: "plugin not found", UpdatedAt: time.Now(), Duration: "0s"}
		m.setResult(res)
		return res
	}

	start := time.Now()
	data, err := p.Run(ctx, req)
	res := Result{
		Name:      name,
		UpdatedAt: time.Now(),
		Duration:  time.Since(start).String(),
	}
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
	} else {
		res.Status = "ok"
		res.Data = data
	}
	m.setResult(res)
	return res
}

func (m *Manager) RunAll(ctx context.Context, req RunRequest) []Result {
	names := m.Names()
	results := make([]Result, 0, len(names))
	var wg sync.WaitGroup
	out := make(chan Result, len(names))

	for _, name := range names {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			out <- m.RunOne(ctx, name, req)
		}()
	}

	wg.Wait()
	close(out)
	for res := range out {
		results = append(results, res)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	return results
}

func (m *Manager) Latest() []Result {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Result, 0, len(m.results))
	for _, res := range m.results {
		out = append(out, res)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (m *Manager) setResult(res Result) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[res.Name] = res
}

type httpPlugin struct{}

func (httpPlugin) Name() string { return "http" }
func (httpPlugin) Description() string {
	return "HTTP status scanner with response time"
}
func (httpPlugin) Run(ctx context.Context, req RunRequest) (any, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target required")
	}
	return scanner.HTTPStatus(ctx, scanner.NormalizeURL(req.Target), req.Timeout), nil
}

type dnsPlugin struct{}

func (dnsPlugin) Name() string { return "dns" }
func (dnsPlugin) Description() string {
	return "DNS lookup for host, IPs and CNAME"
}
func (dnsPlugin) Run(ctx context.Context, req RunRequest) (any, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target required")
	}
	return scanner.DNSLookup(ctx, req.Target), nil
}

type latencyPlugin struct{}

func (latencyPlugin) Name() string { return "latency" }
func (latencyPlugin) Description() string {
	return "TCP latency check"
}
func (latencyPlugin) Run(_ context.Context, req RunRequest) (any, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target required")
	}
	return scanner.Latency(req.Target, 443, req.Timeout), nil
}

type sslPlugin struct{}

func (sslPlugin) Name() string { return "ssl" }
func (sslPlugin) Description() string {
	return "SSL certificate information"
}
func (sslPlugin) Run(_ context.Context, req RunRequest) (any, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target required")
	}
	return scanner.SSLInfo(req.Target, 443, req.Timeout), nil
}

type portsPlugin struct{}

func (portsPlugin) Name() string { return "ports" }
func (portsPlugin) Description() string {
	return "TCP port scan with banner grabbing"
}
func (portsPlugin) Run(ctx context.Context, req RunRequest) (any, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target required")
	}
	if req.StartPort < 1 || req.EndPort < req.StartPort {
		return nil, fmt.Errorf("invalid port range")
	}
	return scanner.PortScan(ctx, req.Target, req.StartPort, req.EndPort, req.Threads, req.Rate, req.Timeout), nil
}

type subdomainsPlugin struct{}

func (subdomainsPlugin) Name() string { return "subdomains" }
func (subdomainsPlugin) Description() string {
	return "Subdomain discovery using internal wordlist"
}
func (subdomainsPlugin) Run(ctx context.Context, req RunRequest) (any, error) {
	if req.Target == "" {
		return nil, fmt.Errorf("target required")
	}
	return subdomain.Find(ctx, req.Target, "", req.Threads, req.Rate, req.Timeout), nil
}

type passwordPlugin struct{}

func (passwordPlugin) Name() string { return "password" }
func (passwordPlugin) Description() string {
	return "Password entropy and strength checker"
}
func (passwordPlugin) Run(_ context.Context, req RunRequest) (any, error) {
	if req.Password == "" {
		return nil, fmt.Errorf("password required")
	}
	return password.Check(req.Password), nil
}

func RegisterDefault(m *Manager) {
	m.Register(httpPlugin{})
	m.Register(portsPlugin{})
	m.Register(subdomainsPlugin{})
	m.Register(dnsPlugin{})
	m.Register(domainMetadataPlugin{})
	m.Register(sslPlugin{})
	m.Register(latencyPlugin{})
	m.Register(passwordPlugin{})
}
