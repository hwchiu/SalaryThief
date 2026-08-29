package events

import (
	"github.com/hwchiu/SalaryThief/internal/model"
	"testing"
)

func TestDedupUsesEventIDThenFallback(t *testing.T) {
	d := NewDeduper()
	e := model.Event{Target: "s", Source: "EventLog", EventID: "1"}
	if !d.Accept(e) || d.Accept(e) {
		t.Fatal("event id dedup")
	}
	a := model.Event{Target: "s", Component: "drive", Location: "Bay 03", EventType: "media_error", Message: "failure"}
	if !d.Accept(a) || d.Accept(a) {
		t.Fatal("fallback dedup")
	}
}
