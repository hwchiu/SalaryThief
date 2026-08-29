package collect

import (
	"context"
	"sync"
	"time"

	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/model"
)

type Runner func(context.Context, config.Target) model.Snapshot
type Scheduler struct {
	targets                              []config.Target
	workers                              int
	interval, initialBackoff, maxBackoff time.Duration
	run                                  Runner
	onActive                             func(int)
	onQueueDepth                         func(int)
}

func (s *Scheduler) SetActivityObserver(observer func(int))   { s.onActive = observer }
func (s *Scheduler) SetQueueDepthObserver(observer func(int)) { s.onQueueDepth = observer }

func NewScheduler(targets []config.Target, workers int, interval, initialBackoff, maxBackoff time.Duration, run Runner) *Scheduler {
	if workers < 1 {
		workers = 1
	}
	return &Scheduler{targets: targets, workers: workers, interval: interval, initialBackoff: initialBackoff, maxBackoff: maxBackoff, run: run}
}

// Run dispatches due targets to a fixed pool; a slow target never forms a global cycle barrier.
func (s *Scheduler) Run(ctx context.Context, result func(model.Snapshot)) {
	jobs := make(chan config.Target, s.workers)
	outcomes := make(chan model.Snapshot, s.workers)
	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-jobs:
					if !ok {
						return
					}
					if s.onQueueDepth != nil {
						s.onQueueDepth(len(jobs))
					}
					if s.onActive != nil {
						s.onActive(1)
					}
					snap := s.run(ctx, t)
					if s.onActive != nil {
						s.onActive(-1)
					}
					select {
					case outcomes <- snap:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	next := map[string]time.Time{}
	failures := map[string]int{}
	tick := time.NewTicker(minDuration(s.interval/4, time.Second))
	defer tick.Stop()
	dispatch := func(now time.Time) {
		for _, t := range s.targets {
			if due := next[t.Name]; !due.IsZero() && now.Before(due) {
				continue
			}
			select {
			case jobs <- t:
				next[t.Name] = now.Add(s.interval)
				if s.onQueueDepth != nil {
					s.onQueueDepth(len(jobs))
				}
			default:
			}
		}
	}
	dispatch(time.Now())
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case now := <-tick.C:
			dispatch(now)
		case snap := <-outcomes:
			result(snap)
			if snap.Up {
				failures[snap.Target] = 0
				next[snap.Target] = time.Now().Add(s.interval)
				dispatch(time.Now())
				continue
			}
			failures[snap.Target]++
			backoff := s.initialBackoff
			for i := 1; i < failures[snap.Target]; i++ {
				backoff *= 2
				if backoff >= s.maxBackoff {
					backoff = s.maxBackoff
					break
				}
			}
			if backoff > s.maxBackoff {
				backoff = s.maxBackoff
			}
			next[snap.Target] = time.Now().Add(backoff)
			dispatch(time.Now())
		}
	}
}
func minDuration(a, b time.Duration) time.Duration {
	if a <= 0 {
		return b
	}
	if a < b {
		return a
	}
	return b
}
