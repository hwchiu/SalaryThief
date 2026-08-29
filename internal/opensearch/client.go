package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/model"
)

// Client uses a bounded asynchronous queue. OpenSearch can never sit on the
// target collection path or the Prometheus scrape path.
type Client struct {
	cfg     config.OpenSearchConfig
	queue   chan []byte
	client  *http.Client
	observe func(int, string)
	mu      sync.RWMutex
}

func New(cfg config.OpenSearchConfig) *Client {
	c := &Client{cfg: cfg, queue: make(chan []byte, 128), client: &http.Client{Timeout: 5 * time.Second}}
	go c.worker()
	return c
}
func (c *Client) SetObserver(observer func(int, string)) {
	c.mu.Lock()
	c.observe = observer
	c.mu.Unlock()
}
func (c *Client) notify(event string) {
	c.mu.RLock()
	observer := c.observe
	c.mu.RUnlock()
	if observer != nil {
		observer(len(c.queue), event)
	}
}
func (c *Client) Publish(_ context.Context, s model.Snapshot) error {
	b, err := json.Marshal(map[string]any{"server": s.Target, "up": s.Up, "last_success": s.LastSuccess, "resources": s.Resources})
	if err != nil {
		c.notify("publish")
		return err
	}
	c.enqueue(b)
	return nil
}
func (c *Client) PublishInventory(_ context.Context, s model.InventorySnapshot) error {
	b, err := json.Marshal(s)
	if err != nil {
		c.notify("publish")
		return err
	}
	c.enqueue(b)
	return nil
}
func (c *Client) enqueue(b []byte) {
	select {
	case c.queue <- b:
		c.notify("")
	default:
		c.notify("dropped")
	}
}
func (c *Client) worker() {
	for b := range c.queue {
		c.notify("")
		if len(c.cfg.Addresses) == 0 {
			c.notify("publish")
			continue
		}
		req, err := http.NewRequest(http.MethodPost, strings.TrimRight(c.cfg.Addresses[0], "/")+"/"+c.cfg.InventoryIndex+"/_doc", bytes.NewReader(b))
		if err != nil {
			c.notify("request")
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if c.cfg.Username != "" {
			req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
		}
		resp, err := c.client.Do(req)
		if err != nil {
			c.notify("request")
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			c.notify("publish")
		}
	}
}
