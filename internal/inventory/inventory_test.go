package inventory

import (
	"github.com/hwchiu/SalaryThief/internal/model"
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
func TestDiffChangesAndMissing(t *testing.T) {
	old := model.InventorySnapshot{Components: []model.InventoryComponent{{Type: "dimm", Location: "A1", Firmware: "1", CapacityBytes: ptr(1)}, {Type: "fan", Location: "F1"}}}
	next := model.InventorySnapshot{Components: []model.InventoryComponent{{Type: "dimm", Location: "A1", Firmware: "2", CapacityBytes: ptr(2)}, {Type: "drive", Location: "Bay 1"}}}
	got := Diff(old, next)
	if len(got) != 4 {
		t.Fatalf("got %d", len(got))
	}
}
