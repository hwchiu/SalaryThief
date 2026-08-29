package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hwchiu/SalaryThief/internal/collect"
	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/inventory"
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
		osClient.SetObserver(reg.ObservePersistence)
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

	var targetGates sync.Map
	acquire := func(ctx context.Context, name string) bool {
		value, _ := targetGates.LoadOrStore(name, make(chan struct{}, 1))
		gate := value.(chan struct{})
		select {
		case gate <- struct{}{}:
			return true
		case <-ctx.Done():
			return false
		}
	}
	release := func(name string) { value, _ := targetGates.Load(name); <-value.(chan struct{}) }
	scheduler := collect.NewScheduler(cfg.Targets, cfg.Workers.Telemetry, cfg.Scheduler.TelemetryInterval, cfg.Retry.InitialBackoff, cfg.Retry.MaxBackoff, func(runCtx context.Context, target config.Target) model.Snapshot {
		if !acquire(runCtx, target.Name) {
			return model.Snapshot{Target: target.Name}
		}
		defer release(target.Name)
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
	// Inventory has its own cadence and worker pool; it never runs on telemetry or scrape paths.
	var inventoryResults sync.Map
	inventoryScheduler := collect.NewScheduler(cfg.Targets, cfg.Workers.Inventory, cfg.Scheduler.InventoryInterval, cfg.Retry.InitialBackoff, cfg.Retry.MaxBackoff, func(runCtx context.Context, target config.Target) model.Snapshot {
		if !acquire(runCtx, target.Name) {
			return model.Snapshot{Target: target.Name}
		}
		defer release(target.Name)
		inv, err := inventory.Collect(runCtx, target)
		if err == nil {
			inventoryResults.Store(target.Name, inv)
		}
		return model.Snapshot{Target: target.Name, Up: err == nil}
	})
	go inventoryScheduler.Run(ctx, func(s model.Snapshot) {
		if s.Up && osClient != nil {
			if value, ok := inventoryResults.LoadAndDelete(s.Target); ok {
				_ = osClient.PublishInventory(context.Background(), value.(model.InventorySnapshot))
			}
		}
	})
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}
