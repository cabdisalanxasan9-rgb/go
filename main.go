package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"dkk/pkg/config"
	"dkk/pkg/logger"
	"dkk/pkg/output"
	"dkk/pkg/scanner"
	"dkk/pkg/utils"
)

const safetyWarning = "Educational and defensive use only. Do not scan systems without explicit authorization."

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	logr, err := logger.New(cfg.Verbose, cfg.Silent, cfg.LogFile)
	if err != nil {
		fmt.Println("failed to initialize logger:", err)
		os.Exit(1)
	}

	if !cfg.Silent {
		fmt.Println("WARNING:", safetyWarning)
	}

	host := ""
	if cfg.Mode != "password" {
		if err := utils.ValidateHost(cfg.Target); err != nil {
			logr.Error("target validation failed", err)
			os.Exit(1)
		}
		host = utils.HostFromTarget(cfg.Target)
	}

	ctx := context.Background()
	result := scanner.ToolkitResult{
		Warning:         safetyWarning,
		Target:          host,
		GeneratedAtUnix: time.Now().Unix(),
	}

	switch cfg.Mode {
	case "http":
		scanHTTP(ctx, cfg, &result, logr)
	case "subdomains":
		scanSubdomains(ctx, cfg, host, &result, logr)
	case "portscan":
		scanPorts(ctx, cfg, host, &result, logr)
	case "password":
		result.Password = scanner.CheckPasswordStrength(cfg.Password)
	case "dns":
		result.DNS = scanner.LookupDNS(ctx, host)
	case "ssl":
		result.SSL = scanner.ReadSSLCert(host, 443, cfg.Timeout)
	case "latency":
		result.Latency = scanner.CheckLatency(host, cfg.LatencyPort, cfg.Timeout)
	case "all":
		scanHTTP(ctx, cfg, &result, logr)
		scanSubdomains(ctx, cfg, host, &result, logr)
		scanPorts(ctx, cfg, host, &result, logr)
		result.DNS = scanner.LookupDNS(ctx, host)
		result.SSL = scanner.ReadSSLCert(host, 443, cfg.Timeout)
		result.Latency = scanner.CheckLatency(host, cfg.LatencyPort, cfg.Timeout)
	}

	if cfg.Mode == "all" || cfg.Mode == "password" {
		if cfg.Password != "" {
			result.Password = scanner.CheckPasswordStrength(cfg.Password)
		}
	}

	if err := output.Write(cfg.Output, cfg.Format, result); err != nil {
		logr.Error("failed to write output", err)
		os.Exit(1)
	}

	if !cfg.Silent {
		fmt.Printf("Done. Mode=%s\n", cfg.Mode)
		if cfg.Output != "" {
			fmt.Printf("Saved output to %s (%s)\n", cfg.Output, cfg.Format)
		}
	}
}

func scanHTTP(ctx context.Context, cfg *config.Config, result *scanner.ToolkitResult, logr *logger.Logger) {
	url := utils.NormalizeURL(cfg.Target)
	logr.Verbose("running HTTP scan", url)
	result.HTTP = scanner.HTTPStatus(ctx, url, cfg.Timeout)
}

func scanSubdomains(ctx context.Context, cfg *config.Config, host string, result *scanner.ToolkitResult, logr *logger.Logger) {
	logr.Verbose("running subdomain scan", host)
	result.Subdomains = scanner.FindSubdomains(ctx, host, cfg.SubdomainWordlist, cfg.Threads, cfg.RateLimit, cfg.Timeout)
}

func scanPorts(ctx context.Context, cfg *config.Config, host string, result *scanner.ToolkitResult, logr *logger.Logger) {
	logr.Verbose("running port scan", host, cfg.StartPort, cfg.EndPort)
	result.Ports = scanner.ScanPorts(ctx, host, cfg.StartPort, cfg.EndPort, cfg.Threads, cfg.RateLimit, cfg.Timeout)
}
