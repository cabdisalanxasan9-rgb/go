package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dkk/pkg/scanner"
	"dkk/pkg/utils"
)

type scanResponse struct {
	Mode      string               `json:"mode"`
	Target    string               `json:"target"`
	Timestamp string               `json:"timestamp"`
	HTTP      *scanner.HTTPResult  `json:"http,omitempty"`
	DNS       *scanner.DNSResult   `json:"dns,omitempty"`
	SSL       *scanner.SSLResult   `json:"ssl,omitempty"`
	Latency   *scanner.LatencyResult `json:"latency,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	targetRaw := strings.TrimSpace(r.URL.Query().Get("target"))
	if targetRaw == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "target query is required"})
		return
	}

	if err := utils.ValidateHost(targetRaw); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid target"})
		return
	}

	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	if mode == "" {
		mode = "http"
	}

	timeoutSeconds := 4
	if raw := strings.TrimSpace(r.URL.Query().Get("timeout")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			if parsed < 1 {
				parsed = 1
			}
			if parsed > 15 {
				parsed = 15
			}
			timeoutSeconds = parsed
		}
	}

	host := utils.HostFromTarget(targetRaw)
	timeout := time.Duration(timeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	resp := scanResponse{
		Mode:      mode,
		Target:    host,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	switch mode {
	case "http":
		resp.HTTP = scanner.HTTPStatus(ctx, utils.NormalizeURL(targetRaw), timeout)
	case "dns":
		resp.DNS = scanner.LookupDNS(ctx, host)
	case "ssl":
		resp.SSL = scanner.ReadSSLCert(host, 443, timeout)
	case "latency":
		resp.Latency = scanner.CheckLatency(host, 443, timeout)
	case "all":
		resp.HTTP = scanner.HTTPStatus(ctx, utils.NormalizeURL(targetRaw), timeout)
		resp.DNS = scanner.LookupDNS(ctx, host)
		resp.SSL = scanner.ReadSSLCert(host, 443, timeout)
		resp.Latency = scanner.CheckLatency(host, 443, timeout)
	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "mode must be one of: http, dns, ssl, latency, all"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
