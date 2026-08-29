package collect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/model"
)

func TestScrapeClassifiesPartialResourceFailure(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redfish/v1/Systems/1/Storage" {
			http.Error(w, "broken", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()
	snap := Scrape(context.Background(), config.Target{Name: "partial", Endpoint: s.URL, Timeout: time.Second, Collect: config.CollectConfig{Thermal: true, Storage: true}}, nil)
	if !snap.Up {
		t.Fatal("core target should stay up")
	}
	if snap.ErrorClass != model.ErrorPartial {
		t.Fatalf("error=%q", snap.ErrorClass)
	}
	if snap.Resources["thermal"].State != model.HealthOK || snap.Resources["storage"].State != model.HealthError {
		t.Fatalf("unexpected resources: %#v", snap.Resources)
	}
}
func TestScrapeClassifiesAuthAndTimeout(t *testing.T) {
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer auth.Close()
	if got := Scrape(context.Background(), config.Target{Name: "auth", Endpoint: auth.URL, Timeout: time.Second}, nil).ErrorClass; got != model.ErrorAuth {
		t.Fatalf("auth=%q", got)
	}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(100 * time.Millisecond) }))
	defer slow.Close()
	if got := Scrape(context.Background(), config.Target{Name: "slow", Endpoint: slow.URL, Timeout: 10 * time.Millisecond}, nil).ErrorClass; got != model.ErrorTimeout {
		t.Fatalf("timeout=%q", got)
	}
}
func TestSchedulerBoundsConcurrency(t *testing.T) {
	targets := make([]config.Target, 8)
	for i := range targets {
		targets[i].Name = string(rune('a' + i))
	}
	var active, max atomic.Int32
	done := make(chan struct{}, 8)
	run := func(_ context.Context, target config.Target) model.Snapshot {
		n := active.Add(1)
		for {
			m := max.Load()
			if n <= m || max.CompareAndSwap(m, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		active.Add(-1)
		return model.Snapshot{Target: target.Name, Up: true}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewScheduler(targets, 2, time.Hour, time.Second, time.Second, run)
	go s.Run(ctx, func(model.Snapshot) { done <- struct{}{} })
	for i := 0; i < len(targets); i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("scheduler did not finish")
		}
	}
	if got := max.Load(); got > 2 {
		t.Fatalf("max active=%d", got)
	}
}

func TestSchedulerNeverCollectsTargetConcurrently(t *testing.T) {
	var active, max atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{}, 2)
	run := func(_ context.Context, target config.Target) model.Snapshot {
		n := active.Add(1)
		for {
			m := max.Load()
			if n <= m || max.CompareAndSwap(m, n) {
				break
			}
		}
		time.Sleep(40 * time.Millisecond)
		active.Add(-1)
		return model.Snapshot{Target: target.Name, Up: true}
	}
	s := NewScheduler([]config.Target{{Name: "one"}}, 2, 5*time.Millisecond, time.Second, time.Second, run)
	go s.Run(ctx, func(model.Snapshot) { done <- struct{}{} })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("no collection")
	}
	time.Sleep(60 * time.Millisecond)
	if got := max.Load(); got != 1 {
		t.Fatalf("max in-flight for one target = %d, want 1", got)
	}
}
