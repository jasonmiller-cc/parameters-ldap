// Command server is the entry point for the parameters-ldap REST API service.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/jasonmiller-cc/parameters-core/pkg/health"
	corelog "github.com/jasonmiller-cc/parameters-core/pkg/log"
	"github.com/jasonmiller-cc/parameters-core/pkg/metrics"
	"github.com/jasonmiller-cc/parameters-core/pkg/middleware"
	"github.com/jasonmiller-cc/parameters-core/pkg/server"
	"github.com/jasonmiller-cc/parameters-ldap/internal/api"
	"github.com/jasonmiller-cc/parameters-ldap/internal/config"
	"github.com/jasonmiller-cc/parameters-ldap/internal/service"
)

const serviceName = "parameters-ldap"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceName, err)
		os.Exit(1)
	}
}

func run() error {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file (default: auto-discover)")
	flag.Parse()

	// Load configuration.
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Build logger.
	logLevel := corelog.LevelInfo
	if cfg.Log.Level == "debug" {
		logLevel = corelog.LevelDebug
	}
	logFormat := corelog.FormatJSON
	if cfg.Log.Format == "text" {
		logFormat = corelog.FormatText
	}
	log := corelog.New(logLevel, logFormat).With("service", serviceName)
	corelog.SetDefault(log)

	// Build LDAP service.
	svc := service.New(cfg)

	// Build health checker with an LDAP ping.
	checker := health.New(serviceName, "dev").
		Add("ldap", svc.Ping)

	// Build Prometheus metrics registry.
	metricsReg := metrics.New("ldap")

	// Build HTTP mux.
	mux := http.NewServeMux()

	// Health / liveness endpoints.
	mux.HandleFunc("GET /healthz", health.LiveHandler())
	mux.HandleFunc("GET /readyz", checker.Handler())

	// Prometheus metrics endpoint.
	metricsPath := "/metrics"
	if cfg.Metrics.Path != "" {
		metricsPath = cfg.Metrics.Path
	}
	mux.Handle("GET "+metricsPath, metricsReg.Handler())

	// API routes.
	handler := api.New(svc)
	handler.Register(mux)

	// Compose middleware chain: metrics → request-id → logging → recovery.
	chain := middleware.Chain(
		metricsReg.Middleware,
		middleware.RequestID,
		middleware.Logger(log),
		middleware.Recover(log),
	)

	// Start server.
	srv := server.New(cfg.Server, chain(mux), log)
	log.Info("starting server",
		"addr", cfg.Server.Addr(),
		"ldap_url", cfg.LDAP.URL,
	)
	return srv.Run(context.Background())
}
