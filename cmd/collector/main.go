package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hwchiu/SalaryThief/internal/collect"
	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/metrics"
	"github.com/hwchiu/SalaryThief/internal/model"
	"github.com/hwchiu/SalaryThief/internal/opensearch"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfgPath := flag.String("config", "configs/collector.yaml", "path to collector config")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Error("load config", "err", err)
		os.Exit(1)
	}

	reg := metrics.New()
	reg.SetTargetCount(len(cfg.Targets))
	var osClient *opensearch.Client
	if cfg.OpenSearch.Enabled {
		osClient = opensearch.New(cfg.OpenSearch)
		log.Info("opensearch enabled", "addresses", cfg.OpenSearch.Addresses)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("listening", "addr", cfg.Listen, "targets", len(cfg.Targets))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server", "err", err)
			stop()
		}
	}()

	scheduler := collect.NewScheduler(cfg.Targets, cfg.Workers.Telemetry, cfg.Scheduler.TelemetryInterval, cfg.Retry.InitialBackoff, cfg.Retry.MaxBackoff, func(runCtx context.Context, target config.Target) model.Snapshot {
		return collect.Scrape(runCtx, target, log)
	})
	var active atomic.Int64
	scheduler.SetActivityObserver(func(delta int) { reg.SetWorkersActive(int(active.Add(int64(delta)))) })
	scheduler.SetQueueDepthObserver(reg.SetQueueDepth)
	go scheduler.Run(ctx, func(snap model.Snapshot) {
		reg.Observe(snap)
		if snap.Up {
			log.Info("scraped", "target", snap.Target, "duration", snap.Duration.String())
		} else {
			log.Warn("scrape failed", "target", snap.Target, "reason", snap.ErrorClass)
		}
		if osClient != nil {
			_ = osClient.Publish(context.Background(), snap)
		}
	})
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}
