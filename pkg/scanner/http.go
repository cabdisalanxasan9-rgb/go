package scanner

import (
	"context"
	"net/http"
	"time"
)

func HTTPStatus(ctx context.Context, target string, timeout time.Duration) *HTTPResult {
	res := &HTTPResult{URL: target}

	client := &http.Client{Timeout: timeout}
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
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
