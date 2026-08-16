package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ctkqiang/argus/internal/config"
	"github.com/ctkqiang/argus/internal/utilities"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "", "Path to YAML configuration file (optional, uses defaults when empty)")
	logLevel := flag.String("log-level", "", "Override log level: debug/info/warn/error (optional)")
	logFormat := flag.String("log-format", "", "Override log format: text/json (optional)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("argus-controller", version)
		os.Exit(0)
	}

	cfg, err := config.LoadControllerFromEnv()
	if err != nil {
		panic("failed to load controller config: " + err.Error())
	}

	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}
	if *logFormat != "" {
		cfg.LogFormat = *logFormat
	}

	logger := utilities.NewLogger(
		utilities.WithLevel(utilities.LogLevel(cfg.LogLevel)),
		utilities.WithFormat(utilities.Format(cfg.LogFormat)),
	)

	if *configPath != "" {
		logger.LogWarn("config file flag specified but YAML loader not yet implemented; using defaults",
			"config_path", *configPath,
		)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	metricsServer := &http.Server{
		Addr:              cfg.MetricsAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.LogInfo("metrics/health server starting", "addr", cfg.MetricsAddr)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError("metrics server error", "err", err.Error())
		}
	}()

	logger.LogInfo("argus-controller starting",
		"version", version,
		"config_path", *configPath,
		"grpc_addr", cfg.GRPCAddr,
		"metrics_addr", cfg.MetricsAddr,
		"leader_election", cfg.EnableLeaderElection,
		"leader_namespace", cfg.LeaderElectionNamespace,
		"event_storage_root", cfg.EventStorageRoot,
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
	)

	logger.LogInfo("argus-controller started, waiting for shutdown signal")

	<-ctx.Done()
	stop()

	logger.LogInfo("received shutdown signal, initiating graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.LogError("metrics server shutdown error", "err", err.Error())
	}

	logger.LogInfo("argus-controller stopped")
}
