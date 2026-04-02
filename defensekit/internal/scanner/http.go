package scanner

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type HTTPResult struct {
	URL          string        `json:"url"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

func NormalizeURL(target string) string {
	t := strings.TrimSpace(target)
	if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
		return t
	}
	return "https://" + t
}

func HTTPStatus(ctx context.Context, url string, timeout time.Duration) *HTTPResult {
	res := &HTTPResult{URL: url}
	client := &http.Client{Timeout: timeout}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		res.Error = err.Error()
		return res
	}

	r, err := client.Do(req)
	res.ResponseTime = time.Since(start)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer r.Body.Close()

	res.StatusCode = r.StatusCode
	return res
}
