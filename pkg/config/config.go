package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Mode              string
	Target            string
	StartPort         int
	EndPort           int
	Threads           int
	Timeout           time.Duration
	RateLimit         int
	Verbose           bool
	Silent            bool
	Output            string
	Format            string
	LogFile           string
	SubdomainWordlist string
	Password          string
	LatencyPort       int
}

func Parse() (*Config, error) {
	cfg := &Config{}
	var timeoutSeconds int

	flag.StringVar(&cfg.Mode, "mode", "all", "Mode: http|subdomains|portscan|password|dns|ssl|latency|all")
	flag.StringVar(&cfg.Target, "target", "", "Target host or domain")
	flag.IntVar(&cfg.StartPort, "start", 1, "Start port for TCP scan")
	flag.IntVar(&cfg.EndPort, "end", 1024, "End port for TCP scan")
	flag.IntVar(&cfg.Threads, "threads", 100, "Worker thread count")
	flag.IntVar(&timeoutSeconds, "timeout", 2, "Timeout in seconds")
	flag.IntVar(&cfg.RateLimit, "rate", 200, "Rate limit (tasks per second)")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&cfg.Silent, "silent", false, "Silent mode (minimal terminal output)")
	flag.StringVar(&cfg.Output, "output", "", "Output file path")
	flag.StringVar(&cfg.Format, "format", "json", "Output format: json|txt")
	flag.StringVar(&cfg.LogFile, "log", "", "Log file path")
	flag.StringVar(&cfg.SubdomainWordlist, "wordlist", "", "External subdomain wordlist file")
	flag.StringVar(&cfg.Password, "password", "", "Password text to evaluate")
	flag.IntVar(&cfg.LatencyPort, "latency-port", 443, "Port for latency check")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Defensive Cybersecurity Toolkit (Go)\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "WARNING: Educational and defensive use only.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Do not scan systems without explicit authorization.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Examples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  go run main.go -mode all -target scanme.nmap.org -start 1 -end 1000 -threads 200 -timeout 1 -rate 100 -output result.json -format json\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  go run main.go -mode subdomains -target example.com -wordlist words.txt -output subs.txt -format txt\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  go run main.go -mode password -password 'StrongP@ssw0rd123'\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg.Mode = strings.ToLower(strings.TrimSpace(cfg.Mode))
	cfg.Format = strings.ToLower(strings.TrimSpace(cfg.Format))
	cfg.Timeout = time.Duration(timeoutSeconds) * time.Second

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	allowedModes := map[string]bool{
		"http": true, "subdomains": true, "portscan": true, "password": true,
		"dns": true, "ssl": true, "latency": true, "all": true,
	}
	if !allowedModes[cfg.Mode] {
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	if cfg.Mode != "password" && strings.TrimSpace(cfg.Target) == "" {
		return fmt.Errorf("target is required for mode %s", cfg.Mode)
	}

	if cfg.StartPort < 1 || cfg.EndPort > 65535 || cfg.StartPort > cfg.EndPort {
		return fmt.Errorf("invalid port range: %d-%d", cfg.StartPort, cfg.EndPort)
	}

	if cfg.Threads < 1 {
		return fmt.Errorf("threads must be >= 1")
	}

	if cfg.RateLimit < 1 {
		return fmt.Errorf("rate must be >= 1")
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be > 0")
	}

	if cfg.Format != "json" && cfg.Format != "txt" {
		return fmt.Errorf("invalid output format: %s", cfg.Format)
	}

	if cfg.Mode == "password" && strings.TrimSpace(cfg.Password) == "" {
		return fmt.Errorf("password cannot be empty in password mode")
	}

	if cfg.SubdomainWordlist != "" {
		if _, err := os.Stat(cfg.SubdomainWordlist); err != nil {
			return fmt.Errorf("cannot access wordlist: %w", err)
		}
	}

	return nil
}
