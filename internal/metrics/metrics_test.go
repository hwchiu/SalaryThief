package metrics

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hwchiu/SalaryThief/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsScrapeDoesNotCallRedfish(t *testing.T) {
	var requests atomic.Int32
	bmc := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests.Add(1) }))
	defer bmc.Close()
	prom := prometheus.NewRegistry()
	cache := NewWithRegisterer(prom)
	cache.Observe(model.Snapshot{Target: "server-1", Scope: "test", Up: true, LastSuccess: time.Now(), Resources: map[string]model.ResourceStatus{"thermal": {State: model.HealthOK, LastSuccess: time.Now()}}})
	h := promhttp.HandlerFor(prom, promhttp.HandlerOpts{})
	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("scrape status=%d", rr.Code)
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("metrics made %d Redfish requests", got)
	}
}

func TestFailedRefreshRetainsLastSuccessAndPartialState(t *testing.T) {
	prom := prometheus.NewRegistry()
	cache := NewWithRegisterer(prom)
	then := time.Now().Add(-time.Minute)
	cache.Observe(model.Snapshot{Target: "server-1", Scope: "test", Up: true, LastSuccess: then, Resources: map[string]model.ResourceStatus{"thermal": {State: model.HealthOK, LastSuccess: then}, "storage": {State: model.HealthOK, LastSuccess: then}}})
	cache.Observe(model.Snapshot{Target: "server-1", Scope: "test", Up: true, ErrorClass: model.ErrorPartial, Resources: map[string]model.ResourceStatus{"storage": {State: model.HealthError, ErrorClass: model.ErrorHTTP}}})
	got, _ := cache.Snapshot("server-1")
	if !got.LastSuccess.Equal(then) {
		t.Fatal("lost last success after partial failure")
	}
	if got.Resources["thermal"].State != model.HealthOK {
		t.Fatal("unrelated resource was erased")
	}
	if got.Resources["storage"].State != model.HealthError {
		t.Fatal("storage error not retained")
	}
}

func TestConnectivityFailureMarksCachedResourcesUnknown(t *testing.T) {
	prom := prometheus.NewRegistry()
	cache := NewWithRegisterer(prom)
	then := time.Now().Add(-time.Minute)
	cache.Observe(model.Snapshot{Target: "server-1", Scope: "test", Up: true, LastSuccess: then, Resources: map[string]model.ResourceStatus{"thermal": {State: model.HealthOK, LastSuccess: then}}})
	cache.Observe(model.Snapshot{Target: "server-1", Scope: "test", Up: false, ErrorClass: model.ErrorNetwork, Resources: map[string]model.ResourceStatus{}})
	got, _ := cache.Snapshot("server-1")
	if got.Up || !got.LastSuccess.Equal(then) {
		t.Fatal("connectivity failure should preserve stale last-success data")
	}
	if got.Resources["thermal"].State != model.HealthUnknown {
		t.Fatalf("resource state=%v, want UNKNOWN", got.Resources["thermal"].State)
	}
}
