package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hwchiu/SalaryThief/internal/model"
	"github.com/prometheus/client_golang/prometheus"
)

type targetState struct{ snapshot model.Snapshot }
type Registry struct {
	mu            sync.RWMutex
	targets       map[string]targetState
	collector     *stateCollector
	errors        *prometheus.CounterVec
	targetCount   atomic.Int64
	workersActive atomic.Int64
	queueDepth    atomic.Int64
}

func (r *Registry) SetTargetCount(n int)   { r.targetCount.Store(int64(n)) }
func (r *Registry) SetWorkersActive(n int) { r.workersActive.Store(int64(n)) }
func (r *Registry) SetQueueDepth(n int)    { r.queueDepth.Store(int64(n)) }

func New() *Registry { return NewWithRegisterer(prometheus.DefaultRegisterer) }
func NewWithRegisterer(registerer prometheus.Registerer) *Registry {
	r := &Registry{targets: map[string]targetState{}}
	r.collector = &stateCollector{r: r}
	r.errors = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "redfish_collection_errors_total", Help: "Redfish collection failures by bounded reason."}, []string{"server", "reason"})
	registerer.MustRegister(r.collector, r.errors)
	return r
}
func (r *Registry) Observe(s model.Snapshot) {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.targets[s.Target]
	if ok {
		if s.LastSuccess.IsZero() {
			s.LastSuccess = old.snapshot.LastSuccess
		}
		if s.Resources == nil {
			s.Resources = old.snapshot.Resources
		} else {
			merged := make(map[string]model.ResourceStatus, len(old.snapshot.Resources)+len(s.Resources))
			for family, previous := range old.snapshot.Resources {
				merged[family] = previous
			}
			for family, next := range s.Resources {
				merged[family] = next
			}
			s.Resources = merged
		}
		for family, next := range s.Resources {
			if prev, exists := old.snapshot.Resources[family]; exists && next.LastSuccess.IsZero() {
				next.LastSuccess = prev.LastSuccess
				s.Resources[family] = next
			}
		}
		// A failed core collection means component health is no longer current.
		// Preserve resource freshness timestamps, but never continue to present
		// previously healthy hardware as currently OK while the BMC is offline.
		if !s.Up {
			for family, status := range s.Resources {
				status.State = model.HealthUnknown
				status.ErrorClass = s.ErrorClass
				s.Resources[family] = status
			}
		}
	}
	if s.Resources == nil {
		s.Resources = map[string]model.ResourceStatus{}
	}
	r.targets[s.Target] = targetState{snapshot: s}
	if s.ErrorClass != model.ErrorNone {
		r.errors.WithLabelValues(s.Target, string(s.ErrorClass)).Inc()
	}
}
func (r *Registry) Snapshot(target string) (model.Snapshot, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.targets[target]
	return s.snapshot, ok
}

type stateCollector struct{ r *Registry }

var (
	upDesc       = prometheus.NewDesc("redfish_up", "Whether core Redfish data is currently reachable (not hardware health).", []string{"server", "observability_scope"}, nil)
	lastDesc     = prometheus.NewDesc("redfish_last_success_timestamp_seconds", "Unix time of last successful core collection.", []string{"server", "observability_scope"}, nil)
	ageDesc      = prometheus.NewDesc("redfish_data_age_seconds", "Seconds since the last successful core collection.", []string{"server", "observability_scope"}, nil)
	resourceDesc = prometheus.NewDesc("redfish_resource_health", "Resource collection health: unknown=0 ok=1 warning=2 critical=3 error=4.", []string{"server", "observability_scope", "resource_family"}, nil)
	targetsDesc  = prometheus.NewDesc("redfish_collector_targets", "Number of configured Redfish targets.", nil, nil)
	workersDesc  = prometheus.NewDesc("redfish_collector_workers_active", "Active collection workers.", []string{"pool"}, nil)
	queueDesc    = prometheus.NewDesc("redfish_collector_queue_depth", "Collection queue depth.", []string{"queue"}, nil)
)

func (c *stateCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- lastDesc
	ch <- ageDesc
	ch <- resourceDesc
	ch <- targetsDesc
	ch <- workersDesc
	ch <- queueDesc
}
func (c *stateCollector) Collect(ch chan<- prometheus.Metric) {
	c.r.mu.RLock()
	defer c.r.mu.RUnlock()
	now := time.Now()
	ch <- prometheus.MustNewConstMetric(targetsDesc, prometheus.GaugeValue, float64(c.r.targetCount.Load()))
	ch <- prometheus.MustNewConstMetric(workersDesc, prometheus.GaugeValue, float64(c.r.workersActive.Load()), "telemetry")
	ch <- prometheus.MustNewConstMetric(queueDesc, prometheus.GaugeValue, float64(c.r.queueDepth.Load()), "telemetry")
	for _, state := range c.r.targets {
		s := state.snapshot
		labels := []string{s.Target, s.Scope}
		up := 0.0
		if s.Up {
			up = 1
		}
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, up, labels...)
		if !s.LastSuccess.IsZero() {
			ch <- prometheus.MustNewConstMetric(lastDesc, prometheus.GaugeValue, float64(s.LastSuccess.Unix()), labels...)
			ch <- prometheus.MustNewConstMetric(ageDesc, prometheus.GaugeValue, now.Sub(s.LastSuccess).Seconds(), labels...)
		}
		for family, status := range s.Resources {
			ch <- prometheus.MustNewConstMetric(resourceDesc, prometheus.GaugeValue, float64(status.State), s.Target, s.Scope, family)
		}
	}
}
