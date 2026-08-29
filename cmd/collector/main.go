package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hwchiu/SalaryThief/internal/collect"
	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/metrics"
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

	runOnce(ctx, cfg, reg, osClient, log)
	ticker := time.NewTicker(cfg.ScrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdown)
			return
		case <-ticker.C:
			runOnce(ctx, cfg, reg, osClient, log)
		}
	}
}

func runOnce(ctx context.Context, cfg *config.Config, reg *metrics.Registry, osClient *opensearch.Client, log *slog.Logger) {
	var wg sync.WaitGroup
	for _, t := range cfg.Targets {
		t := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			sctx, cancel := context.WithTimeout(ctx, t.Timeout+2*time.Second)
			defer cancel()
			snap := collect.Scrape(sctx, t, log)
			reg.Observe(snap)
			if snap.Up {
				log.Info("scraped", "target", t.Name, "duration", snap.Duration.String(), "systems", len(snap.Systems), "chassis", len(snap.Chassis))
			} else {
				log.Warn("scrape failed", "target", t.Name, "err", snap.Error)
			}
			if osClient != nil {
				if err := osClient.Publish(sctx, snap); err != nil {
					log.Warn("opensearch publish failed", "target", t.Name, "err", err)
				}
			}
		}()
	}
	wg.Wait()
}
