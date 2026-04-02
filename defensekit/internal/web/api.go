package web

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"defensekit/internal/logger"
	"defensekit/internal/password"
	"defensekit/internal/scanner"
	"defensekit/internal/subdomain"
	"defensekit/internal/web/plugins"
)

type apiConfig struct {
	DefaultThreads int
	DefaultTimeout time.Duration
	DefaultRate    int
	DefaultStart   int
	DefaultEnd     int
}

type runPluginRequest struct {
	Plugin         string `json:"plugin"`
	Target         string `json:"target"`
	Password       string `json:"password"`
	Threads        int    `json:"threads"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Rate           int    `json:"rate"`
	StartPort      int    `json:"start_port"`
	EndPort        int    `json:"end_port"`
	RunAll         bool   `json:"run_all"`
}

type scanRequest struct {
	Mode     string `json:"mode"`
	Target   string `json:"target"`
	Password string `json:"password"`
}

func RegisterAPI(mux *http.ServeMux, warning string, log *logger.Logger, mgr *plugins.Manager, cfg apiConfig) {
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "warning": warning})
	})

	mux.HandleFunc("/api/http", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "target query is required", http.StatusBadRequest)
			return
		}
		res := scanner.HTTPStatus(r.Context(), scanner.NormalizeURL(target), 3*time.Second)
		log.Info("api http scan", target)
		_ = json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("/api/plugins/list", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"plugins":      mgr.Names(),
			"descriptions": mgr.Describe(),
		})
	})

	mux.HandleFunc("/api/plugins/results", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"results": mgr.Latest()})
	})

	mux.HandleFunc("/api/plugins/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		var req runPluginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		runReq := plugins.RunRequest{
			Target:    normalizeTarget(req.Target),
			Threads:   fallbackInt(req.Threads, cfg.DefaultThreads),
			Timeout:   time.Duration(fallbackInt(req.TimeoutSeconds, int(cfg.DefaultTimeout.Seconds()))) * time.Second,
			Rate:      fallbackInt(req.Rate, cfg.DefaultRate),
			StartPort: fallbackInt(req.StartPort, cfg.DefaultStart),
			EndPort:   fallbackInt(req.EndPort, cfg.DefaultEnd),
			Password:  req.Password,
		}

		if runReq.Target == "" && runReq.Password == "" {
			http.Error(w, "target or password is required", http.StatusBadRequest)
			return
		}

		if req.RunAll {
			results := mgr.RunAll(r.Context(), runReq)
			log.Info("api plugin run_all", runReq.Target)
			_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
			return
		}

		pluginName := strings.TrimSpace(req.Plugin)
		if pluginName == "" {
			http.Error(w, "plugin is required when run_all=false", http.StatusBadRequest)
			return
		}

		res := mgr.RunOne(r.Context(), pluginName, runReq)
		log.Info("api plugin run", pluginName, runReq.Target)
		_ = json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		var req scanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		mode := strings.ToLower(strings.TrimSpace(req.Mode))
		target := normalizeTarget(req.Target)
		timeout := cfg.DefaultTimeout
		ctx := context.Background()

		result := map[string]any{
			"mode":    mode,
			"target":  target,
			"warning": warning,
			"time":    time.Now().UTC().Format(time.RFC3339),
		}

		switch mode {
		case "http":
			result["http"] = scanner.HTTPStatus(ctx, scanner.NormalizeURL(target), timeout)
		case "portscan":
			result["ports"] = scanner.PortScan(ctx, target, cfg.DefaultStart, cfg.DefaultEnd, cfg.DefaultThreads, cfg.DefaultRate, timeout)
		case "subdomains":
			result["subdomains"] = subdomain.Find(ctx, target, "", cfg.DefaultThreads, cfg.DefaultRate, timeout)
		case "dns":
			result["dns"] = scanner.DNSLookup(ctx, target)
		case "ssl":
			result["ssl"] = scanner.SSLInfo(target, 443, timeout)
		case "latency":
			result["latency"] = scanner.Latency(target, 443, timeout)
		case "password":
			if strings.TrimSpace(req.Password) == "" {
				http.Error(w, "password is required for mode=password", http.StatusBadRequest)
				return
			}
			result["password"] = password.Check(req.Password)
		case "all":
			result["http"] = scanner.HTTPStatus(ctx, scanner.NormalizeURL(target), timeout)
			result["ports"] = scanner.PortScan(ctx, target, cfg.DefaultStart, cfg.DefaultEnd, cfg.DefaultThreads, cfg.DefaultRate, timeout)
			result["subdomains"] = subdomain.Find(ctx, target, "", cfg.DefaultThreads, cfg.DefaultRate, timeout)
			result["dns"] = scanner.DNSLookup(ctx, target)
			result["ssl"] = scanner.SSLInfo(target, 443, timeout)
			result["latency"] = scanner.Latency(target, 443, timeout)
			if strings.TrimSpace(req.Password) != "" {
				result["password"] = password.Check(req.Password)
			}
		default:
			http.Error(w, "invalid mode", http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(result)
	})
}

func fallbackInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func normalizeTarget(raw string) string {
	t := strings.TrimSpace(raw)
	if strings.Contains(t, "://") {
		if host, _, err := net.SplitHostPort(t); err == nil && host != "" {
			return host
		}
		t = strings.TrimPrefix(strings.TrimPrefix(t, "http://"), "https://")
		t = strings.Split(t, "/")[0]
		t = strings.Split(t, ":")[0]
	}
	if strings.Contains(t, "/") {
		t = strings.Split(t, "/")[0]
	}
	if strings.Contains(t, ":") {
		parts := strings.Split(t, ":")
		if len(parts) > 1 {
			if _, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				return strings.Join(parts[:len(parts)-1], ":")
			}
		}
	}
	return t
}
