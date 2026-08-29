package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/model"
)

// Client uses a bounded asynchronous queue. OpenSearch can never sit on the
// target collection path or the Prometheus scrape path.
type Client struct {
	cfg    config.OpenSearchConfig
	queue  chan model.Snapshot
	client *http.Client
}

func New(cfg config.OpenSearchConfig) *Client {
	c := &Client{cfg: cfg, queue: make(chan model.Snapshot, 128), client: &http.Client{}}
	go c.worker()
	return c
}
func (c *Client) Publish(_ context.Context, s model.Snapshot) error {
	select {
	case c.queue <- s:
	default:
	}
	return nil
}
func (c *Client) worker() {
	for s := range c.queue {
		if len(c.cfg.Addresses) == 0 {
			continue
		}
		b, _ := json.Marshal(map[string]any{"server": s.Target, "up": s.Up, "last_success": s.LastSuccess, "resources": s.Resources})
		req, err := http.NewRequest(http.MethodPost, strings.TrimRight(c.cfg.Addresses[0], "/")+"/"+c.cfg.InventoryIndex+"/_doc", bytes.NewReader(b))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if c.cfg.Username != "" {
			req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
		}
		resp, err := c.client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}
}
