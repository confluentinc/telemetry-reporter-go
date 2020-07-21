// Package collect implements a library to create
// and define different pull based collectors.
package collect

import (
	"go.opencensus.io/stats"
)

// Collector is an interface that collectors implement
// to collect data.
type Collector interface {
	Collect()
}

// CollectorAgent defines the wrapper format of Collectors
// and data needed by all general collectors.
type CollectorAgent struct {
	Collector
	Config *Config
}

// Config defines the data format of the general
// configurations of a collector.
type Config struct {
	IncludeFilter  string
	StatsToCollect []*stats.Float64Measure
}

// NewConfig returns a pointer to a new collector Config.
func NewConfig(filter string) *Config {
	return &Config{
		IncludeFilter:  filter,
		StatsToCollect: []*stats.Float64Measure{},
	}
}

// NewCollectorAgent returns a pointer to a new CollectorAgent.
func NewCollectorAgent(c Collector, config *Config) *CollectorAgent {
	return &CollectorAgent{
		Collector: c,
		Config:    config,
	}
}
