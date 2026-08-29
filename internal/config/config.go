package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen         string           `yaml:"listen"`
	ScrapeInterval time.Duration    `yaml:"scrape_interval"`
	Scheduler      SchedulerConfig  `yaml:"scheduler"`
	Workers        WorkersConfig    `yaml:"workers"`
	Retry          RetryConfig      `yaml:"retry"`
	OpenSearch     OpenSearchConfig `yaml:"opensearch"`
	Targets        []Target         `yaml:"targets"`
}
type SchedulerConfig struct {
	TelemetryInterval time.Duration `yaml:"telemetry_interval"`
	InventoryInterval time.Duration `yaml:"inventory_interval"`
}
type WorkersConfig struct {
	Telemetry int `yaml:"telemetry"`
	Inventory int `yaml:"inventory"`
}
type RetryConfig struct {
	InitialBackoff time.Duration `yaml:"initial_backoff"`
	MaxBackoff     time.Duration `yaml:"max_backoff"`
}
type OpenSearchConfig struct {
	Enabled            bool     `yaml:"enabled"`
	Addresses          []string `yaml:"addresses"`
	Username           string   `yaml:"username"`
	Password           string   `yaml:"password"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify"`
	InventoryIndex     string   `yaml:"inventory_index"`
	EventsIndex        string   `yaml:"events_index"`
}
type Target struct {
	Name               string            `yaml:"name"`
	Endpoint           string            `yaml:"endpoint"`
	Username           string            `yaml:"username"`
	Password           string            `yaml:"password"`
	InsecureSkipVerify bool              `yaml:"insecure_skip_verify"`
	Vendor             string            `yaml:"vendor"`
	Timeout            time.Duration     `yaml:"timeout"`
	Labels             map[string]string `yaml:"labels"`
	Collect            CollectConfig     `yaml:"collect"`
}
type CollectConfig struct {
	Systems  bool `yaml:"systems"`
	Chassis  bool `yaml:"chassis"`
	Managers bool `yaml:"managers"`
	Thermal  bool `yaml:"thermal"`
	Power    bool `yaml:"power"`
	Storage  bool `yaml:"storage"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b = []byte(os.ExpandEnv(string(b)))
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.Listen == "" {
		c.Listen = ":9100"
	}
	if c.Scheduler.TelemetryInterval <= 0 {
		c.Scheduler.TelemetryInterval = c.ScrapeInterval
	}
	if c.Scheduler.TelemetryInterval <= 0 {
		c.Scheduler.TelemetryInterval = 30 * time.Second
	}
	if c.Workers.Telemetry <= 0 {
		c.Workers.Telemetry = 4
	}
	if c.Workers.Inventory <= 0 {
		c.Workers.Inventory = 1
	}
	if c.Scheduler.InventoryInterval <= 0 {
		c.Scheduler.InventoryInterval = 6 * time.Hour
	}
	if c.Retry.InitialBackoff <= 0 {
		c.Retry.InitialBackoff = 2 * time.Second
	}
	if c.Retry.MaxBackoff <= 0 {
		c.Retry.MaxBackoff = time.Minute
	}
	for i := range c.Targets {
		if c.Targets[i].Timeout <= 0 {
			c.Targets[i].Timeout = 10 * time.Second
		}
	}
	return &c, nil
}
