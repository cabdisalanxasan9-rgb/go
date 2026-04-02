package plugins

import (
	"context"
	"strings"
	"time"
)

type examplePlugin struct{}

func (examplePlugin) Name() string { return "example_plugin" }

func (examplePlugin) Description() string {
	return "Example plugin showing modular extension points"
}

func (examplePlugin) Run(_ context.Context, req RunRequest) (any, error) {
	return map[string]any{
		"message": "example plugin executed",
		"target":  req.Target,
		"has_dot": strings.Contains(req.Target, "."),
		"time":    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func RegisterExample(m *Manager) {
	m.Register(examplePlugin{})
}
