package web

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"defensekit/internal/logger"
	"defensekit/internal/web/plugins"
)

//go:embed templates/*
var templatesFS embed.FS

type ServerConfig struct {
	Threads   int
	Timeout   time.Duration
	Rate      int
	StartPort int
	EndPort   int
}

func StartServer(addr, warning string, log *logger.Logger, cfg ServerConfig) error {
	mux := http.NewServeMux()
	pluginManager := plugins.NewManager()
	plugins.RegisterDefault(pluginManager)
	plugins.RegisterExample(pluginManager)

	dashboardTpl, err := template.ParseFS(templatesFS, "templates/dashboard.html")
	if err != nil {
		return err
	}

	liveTpl, err := template.ParseFS(templatesFS, "templates/live_updates.html")
	if err != nil {
		return err
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_ = dashboardTpl.Execute(w, map[string]any{
			"Title":   "DefenseKit Dashboard",
			"Warning": warning,
		})
	})

	mux.HandleFunc("/live", func(w http.ResponseWriter, _ *http.Request) {
		_ = liveTpl.Execute(w, map[string]any{
			"Title":   "DefenseKit Live Updates",
			"Warning": warning,
		})
	})

	RegisterAPI(mux, warning, log, pluginManager, apiConfig{
		DefaultThreads: fallbackInt(cfg.Threads, 100),
		DefaultTimeout: defaultDuration(cfg.Timeout, 3*time.Second),
		DefaultRate:    fallbackInt(cfg.Rate, 200),
		DefaultStart:   fallbackInt(cfg.StartPort, 1),
		DefaultEnd:     fallbackInt(cfg.EndPort, 1024),
	})

	log.Info("starting web server", addr)
	return http.ListenAndServe(addr, mux)
}

func defaultDuration(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}
