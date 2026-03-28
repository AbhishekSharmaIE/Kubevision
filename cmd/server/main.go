package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api"
	"github.com/AbhishekSharmaIE/Kubevision/internal/auth"
	"github.com/AbhishekSharmaIE/Kubevision/internal/db"
	"github.com/AbhishekSharmaIE/Kubevision/internal/k8s"
	"github.com/AbhishekSharmaIE/Kubevision/internal/metrics"
)

// Version is set at link time; defaults to dev.
var Version = "dev"

func main() {
	_ = os.Setenv("APP_VERSION", Version)

	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("starting KubeVision", "version", Version)

	pgPool := db.MustConnectPostgres()
	defer pgPool.Close()

	redisClient := db.MustConnectRedis()
	defer redisClient.Close()

	if err := db.RunMigrations(context.Background(), pgPool); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	clusterManager := k8s.NewClusterManager(pgPool)
	if err := clusterManager.LoadClusters(context.Background()); err != nil {
		slog.Warn("could not load clusters from DB", "error", err)
	}

	collector := metrics.NewCollector(clusterManager, redisClient)
	ctxCollector, cancelCollector := context.WithCancel(context.Background())
	defer cancelCollector()
	go collector.Start(ctxCollector)

	jwtSvc, err := auth.NewJWTFromEnv()
	if err != nil {
		slog.Error("jwt configuration invalid", "error", err)
		os.Exit(1)
	}

	router := api.NewRouter(api.RouterConfig{
		ClusterManager:   clusterManager,
		MetricsCollector: collector,
		DB:               pgPool,
		Redis:            redisClient,
		JWT:              jwtSvc,
	})

	srv := &http.Server{
		Addr:         api.ListenAddr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown", "error", err)
	}
	slog.Info("server stopped")
}
