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

// Config defines the data format of the general
// configurations of a collector.
type Config struct {
	IncludeFilter  string
	StatsToCollect []*stats.Float64Measure
}

// NewConfig returns a pointer to a new collector Config.
func NewConfig(filter string) Config {
	return Config{
		IncludeFilter:  filter,
		StatsToCollect: []*stats.Float64Measure{},
	}
}
