package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"defensekit/internal/logger"
	"defensekit/internal/output"
	"defensekit/internal/password"
	"defensekit/internal/scanner"
	"defensekit/internal/subdomain"
	"defensekit/internal/web"
)

const warningText = "Educational and defensive use only. Do not scan systems without explicit authorization."

type result struct {
	Warning    string                 `json:"warning"`
	Target     string                 `json:"target,omitempty"`
	HTTP       *scanner.HTTPResult    `json:"http,omitempty"`
	Ports      []scanner.PortResult   `json:"ports,omitempty"`
	Subdomains []subdomain.Result     `json:"subdomains,omitempty"`
	Password   *password.Result       `json:"password,omitempty"`
	DNS        *scanner.DNSResult     `json:"dns,omitempty"`
	SSL        *scanner.SSLResult     `json:"ssl,omitempty"`
	Latency    *scanner.LatencyResult `json:"latency,omitempty"`
	When       int64                  `json:"generated_at_unix"`
}

func main() {
	mode := flag.String("mode", envOrDefault("DK_MODE", "all"), "Mode: http|subdomains|portscan|password|dns|ssl|latency|all")
	target := flag.String("target", envOrDefault("DK_TARGET", ""), "Target host or domain")
	start := flag.Int("start", envIntOrDefault("DK_START_PORT", 1), "Start port")
	end := flag.Int("end", envIntOrDefault("DK_END_PORT", 1024), "End port")
	threads := flag.Int("threads", envIntOrDefault("DK_THREADS", 100), "Worker count")
	timeout := flag.Int("timeout", envIntOrDefault("DK_TIMEOUT", 2), "Timeout in seconds")
	rate := flag.Int("rate", envIntOrDefault("DK_RATE", 200), "Rate limit tasks/sec")
	wordlist := flag.String("wordlist", envOrDefault("DK_WORDLIST", ""), "External subdomain wordlist file")
	pass := flag.String("password", envOrDefault("DK_PASSWORD", ""), "Password to evaluate")
	outputFile := flag.String("output", envOrDefault("DK_OUTPUT", ""), "Output file path")
	format := flag.String("format", envOrDefault("DK_FORMAT", "json"), "Output format: json|txt")
	logFile := flag.String("log", envOrDefault("DK_LOG", ""), "Log file path")
	verbose := flag.Bool("verbose", envBoolOrDefault("DK_VERBOSE", false), "Verbose output")
	silent := flag.Bool("silent", envBoolOrDefault("DK_SILENT", false), "Silent output")
	latencyPort := flag.Int("latency-port", envIntOrDefault("DK_LATENCY_PORT", 443), "Port for latency check")
	serve := flag.Bool("serve", envBoolOrDefault("DK_SERVE", false), "Start REST API and dashboard")
	addr := flag.String("addr", envOrDefault("DK_ADDR", ":8080"), "Web server address")
	healthcheck := flag.Bool("healthcheck", false, "Run HTTP healthcheck and exit")
	healthURL := flag.String("health-url", envOrDefault("DK_HEALTH_URL", "http://127.0.0.1:8080/api/health"), "Healthcheck URL")
	healthTimeout := flag.Int("health-timeout", envIntOrDefault("DK_HEALTH_TIMEOUT", 3), "Healthcheck timeout in seconds")

	flag.Usage = func() {
		fmt.Println("DefenseKit - Defensive Cybersecurity Toolkit")
		fmt.Println(warningText)
		fmt.Println("Examples:")
		fmt.Println("  go run ./cmd -mode all -target scanme.nmap.org -start 1 -end 1000 -threads 200 -timeout 1 -rate 100 -output result.json")
		fmt.Println("  go run ./cmd -mode password -password 'StrongP@ssw0rd123'")
		fmt.Println("  go run ./cmd -serve -addr :8080")
		fmt.Println("  open / for dashboard and /live for live plugin updates")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *healthcheck {
		if err := runHealthcheck(*healthURL, time.Duration(*healthTimeout)*time.Second); err != nil {
			fmt.Println("healthcheck failed:", err)
			os.Exit(1)
		}
		if !*silent {
			fmt.Println("healthcheck ok")
		}
		return
	}

	logr, err := logger.New(*verbose, *silent, *logFile)
	if err != nil {
		fmt.Println("logger init failed:", err)
		os.Exit(1)
	}

	if *serve {
		if !*silent {
			fmt.Println("WARNING:", warningText)
		}
		if err := web.StartServer(*addr, warningText, logr, web.ServerConfig{
			Threads:   *threads,
			Timeout:   time.Duration(*timeout) * time.Second,
			Rate:      *rate,
			StartPort: *start,
			EndPort:   *end,
		}); err != nil {
			logr.Error("web server failed", err)
			os.Exit(1)
		}
		return
	}

	currentMode := strings.ToLower(strings.TrimSpace(*mode))
	if currentMode != "password" && strings.TrimSpace(*target) == "" {
		fmt.Println("target is required")
		os.Exit(1)
	}

	if currentMode != "password" {
		normalized, err := validateTarget(*target)
		if err != nil {
			fmt.Println("invalid target:", err)
			os.Exit(1)
		}
		*target = normalized
	}

	timeoutDur := time.Duration(*timeout) * time.Second
	ctx := context.Background()

	out := result{
		Warning: warningText,
		Target:  *target,
		When:    time.Now().Unix(),
	}

	if !*silent {
		fmt.Println("WARNING:", warningText)
	}

	switch currentMode {
	case "http":
		out.HTTP = scanner.HTTPStatus(ctx, scanner.NormalizeURL(*target), timeoutDur)
	case "subdomains":
		out.Subdomains = subdomain.Find(ctx, *target, *wordlist, *threads, *rate, timeoutDur)
	case "portscan":
		out.Ports = scanner.PortScan(ctx, *target, *start, *end, *threads, *rate, timeoutDur)
	case "password":
		if strings.TrimSpace(*pass) == "" {
			fmt.Println("password is required in password mode")
			os.Exit(1)
		}
		out.Password = password.Check(*pass)
	case "dns":
		out.DNS = scanner.DNSLookup(ctx, *target)
	case "ssl":
		out.SSL = scanner.SSLInfo(*target, 443, timeoutDur)
	case "latency":
		out.Latency = scanner.Latency(*target, *latencyPort, timeoutDur)
	case "all":
		out.HTTP = scanner.HTTPStatus(ctx, scanner.NormalizeURL(*target), timeoutDur)
		out.Subdomains = subdomain.Find(ctx, *target, *wordlist, *threads, *rate, timeoutDur)
		out.Ports = scanner.PortScan(ctx, *target, *start, *end, *threads, *rate, timeoutDur)
		out.DNS = scanner.DNSLookup(ctx, *target)
		out.SSL = scanner.SSLInfo(*target, 443, timeoutDur)
		out.Latency = scanner.Latency(*target, *latencyPort, timeoutDur)
		if strings.TrimSpace(*pass) != "" {
			out.Password = password.Check(*pass)
		}
	default:
		fmt.Println("invalid mode")
		os.Exit(1)
	}

	if err := output.Write(*outputFile, *format, out); err != nil {
		logr.Error("write output failed", err)
		os.Exit(1)
	}

	if !*silent {
		fmt.Printf("Done. mode=%s\n", currentMode)
		if *outputFile != "" {
			fmt.Printf("Output saved: %s\n", *outputFile)
		}
	}
}

func runHealthcheck(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envIntOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func validateTarget(raw string) (string, error) {
	target := strings.TrimSpace(raw)
	if target == "" {
		return "", fmt.Errorf("target is empty")
	}

	if strings.Contains(target, "://") {
		u, err := url.Parse(target)
		if err != nil {
			return "", err
		}
		target = u.Hostname()
	}

	target = strings.TrimSuffix(target, "/")
	target = strings.TrimPrefix(strings.TrimPrefix(target, "http://"), "https://")
	target = strings.Split(target, "/")[0]
	if host, _, err := net.SplitHostPort(target); err == nil {
		target = host
	}

	if net.ParseIP(target) != nil {
		return target, nil
	}

	if !strings.Contains(target, ".") || len(target) < 3 {
		return "", fmt.Errorf("host/domain looks invalid")
	}

	return target, nil
}
