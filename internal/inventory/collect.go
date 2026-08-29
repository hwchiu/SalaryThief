package inventory

import (
	"context"
	"encoding/json"
	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/model"
	"net/http"
	"strings"
	"time"
)

func Collect(ctx context.Context, t config.Target) (model.InventorySnapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(t.Endpoint, "/")+"/redfish/v1/Systems/1/Inventory", nil)
	if err != nil {
		return model.InventorySnapshot{}, err
	}
	if t.Username != "" {
		req.SetBasicAuth(t.Username, t.Password)
	}
	c := &http.Client{Timeout: t.Timeout}
	resp, err := c.Do(req)
	if err != nil {
		return model.InventorySnapshot{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return model.InventorySnapshot{}, &statusError{resp.StatusCode}
	}
	var s model.InventorySnapshot
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return model.InventorySnapshot{}, err
	}
	s.Target = t.Name
	s.Scope = t.Labels["observability_scope"]
	if s.Scope == "" {
		s.Scope = "default"
	}
	s.ObservedAt = time.Now()
	return s, nil
}

type statusError struct{ code int }

func (e *statusError) Error() string { return "inventory HTTP error" }
