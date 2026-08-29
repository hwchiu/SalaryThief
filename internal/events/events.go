package events

import (
	"crypto/sha256"
	"fmt"
	"github.com/hwchiu/SalaryThief/internal/model"
	"sync"
)

type Deduper struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func NewDeduper() *Deduper { return &Deduper{seen: map[string]struct{}{}} }
func Key(e model.Event) string {
	if e.EventID != "" {
		return e.Target + "|" + e.Source + "|" + e.EventID
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%s", e.Target, e.Component, e.Location, e.EventType, e.Message)))
	return fmt.Sprintf("%x", sum)
}
func (d *Deduper) Accept(e model.Event) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	k := Key(e)
	if _, ok := d.seen[k]; ok {
		return false
	}
	d.seen[k] = struct{}{}
	return true
}
