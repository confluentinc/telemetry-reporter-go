package collect

import (
	"go.opencensus.io/stats"
)

// Collector ...
type Collector interface {
	NewCollector()
	Collect()
}

// CollectorAgent ...
type CollectorAgent struct {
	Collector
	Config *Config
}

// Config ...
type Config struct {
	IncludeFilter  string
	StatsToCollect []*stats.Float64Measure
}

// NewConfig ...
func NewConfig(filter string) *Config {
	return &Config{
		IncludeFilter:  filter,
		StatsToCollect: []*stats.Float64Measure{},
	}
}

// NewCollectorAgent ...
func NewCollectorAgent(c Collector, config *Config) *CollectorAgent {
	return &CollectorAgent{
		Collector: c,
		Config:    config,
	}
}
