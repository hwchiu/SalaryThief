package inventory

import (
	"context"
	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func ptr(v uint64) *uint64 { return &v }
func TestDiffReplacementKeepsBay(t *testing.T) {
	old := model.InventorySnapshot{Target: "s", Components: []model.InventoryComponent{{Type: "drive", ComponentID: "drive-03", Location: "Bay 03", Serial: "AAA"}}}
	next := old
	next.ObservedAt = time.Now()
	next.Components = []model.InventoryComponent{{Type: "drive", ComponentID: "drive-03", Location: "Bay 03", Serial: "BBB"}}
	got := Diff(old, next)
	if len(got) != 1 || got[0].Change != "replaced" || got[0].Location != "Bay 03" || got[0].OldSerial != "AAA" || got[0].NewSerial != "BBB" {
		t.Fatalf("%+v", got)
	}
}
func TestCollectNormalizesOptionalInventory(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"components":[{"type":"drive","component_id":"d3","location":"Bay 03","serial":"AAA"},{"type":"fan","component_id":"f1","location":"Fan 1"}]}`))
	}))
	defer s.Close()
	got, err := Collect(context.Background(), config.Target{Name: "server", Endpoint: s.URL, Timeout: time.Second})
	if err != nil || len(got.Components) != 2 || got.Components[1].Serial != "" {
		t.Fatalf("%+v %v", got, err)
	}
}
func TestCollectPartialFailure(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", 500) }))
	defer s.Close()
	_, err := Collect(context.Background(), config.Target{Name: "server", Endpoint: s.URL, Timeout: time.Second})
	if err == nil {
		t.Fatal("want error")
	}
}
func TestDiffChangesAndMissing(t *testing.T) {
	old := model.InventorySnapshot{Components: []model.InventoryComponent{{Type: "dimm", Location: "A1", Firmware: "1", CapacityBytes: ptr(1)}, {Type: "fan", Location: "F1"}}}
	next := model.InventorySnapshot{Components: []model.InventoryComponent{{Type: "dimm", Location: "A1", Firmware: "2", CapacityBytes: ptr(2)}, {Type: "drive", Location: "Bay 1"}}}
	got := Diff(old, next)
	if len(got) != 4 {
		t.Fatalf("got %d", len(got))
	}
}
