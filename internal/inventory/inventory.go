package inventory

import (
	"github.com/hwchiu/SalaryThief/internal/model"
	"sort"
)

func key(c model.InventoryComponent) string { return c.Type + "|" + c.Location }

// Diff compares physical topology by type/location, never by serial number.
func Diff(old, next model.InventorySnapshot) []model.InventoryChange {
	before := map[string]model.InventoryComponent{}
	after := map[string]model.InventoryComponent{}
	for _, c := range old.Components {
		before[key(c)] = c
	}
	for _, c := range next.Components {
		after[key(c)] = c
	}
	out := []model.InventoryChange{}
	for k, b := range before {
		a, ok := after[k]
		if !ok {
			out = append(out, change(next, b, model.InventoryComponent{}, "removed"))
			continue
		}
		if b.Serial != "" && a.Serial != "" && b.Serial != a.Serial {
			out = append(out, change(next, b, a, "replaced"))
			continue
		}
		if b.Firmware != a.Firmware {
			out = append(out, change(next, b, a, "firmware_changed"))
		}
		if b.Model != a.Model {
			out = append(out, change(next, b, a, "model_changed"))
		}
		if !sameCapacity(b.CapacityBytes, a.CapacityBytes) {
			out = append(out, change(next, b, a, "capacity_changed"))
		}
	}
	for k, a := range after {
		if _, ok := before[k]; !ok {
			out = append(out, change(next, model.InventoryComponent{}, a, "added"))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Location < out[j].Location })
	return out
}
func sameCapacity(a, b *uint64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
func change(s model.InventorySnapshot, b, a model.InventoryComponent, kind string) model.InventoryChange {
	c := a
	if c.Location == "" {
		c = b
	}
	return model.InventoryChange{Target: s.Target, Type: c.Type, ComponentID: c.ComponentID, Location: c.Location, Change: kind, OldSerial: b.Serial, NewSerial: a.Serial, ObservedAt: s.ObservedAt}
}
