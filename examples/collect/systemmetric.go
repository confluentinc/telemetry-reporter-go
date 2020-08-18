// Package example shows an example of how to define
// your own custom collector. This example will collect one metric,
// allocated bytes, from the runtime package.
package example

import (
	"context"
	"runtime"
	"time"

	"github.com/confluentinc/telemetry-reporter-go/collect"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// First you must define the OpenCensus Measures of metrics you want to collect
// and add them to a list of measure pointers
var (
	alloc = stats.Int64("alloc", "bytes of allocated heap objects", "by")

	statsList = []*stats.Int64Measure{
		alloc,
	}
)

// Then you define the OpenCensus View for the metric and add it to
// a list of view pointers
var (
	allocView = &view.View{
		Name:        "allocView",
		Measure:     alloc,
		Description: "view for allocated bytes",
		Aggregation: view.LastValue(),
	}

	viewList = []*view.View{
		allocView,
	}
)

// SystemMetric is a collector that collects specific
// system metrics such as CPU utilization.
type SystemMetric struct {
	memStats runtime.MemStats
	config   collect.Config
}

// NewSystemMetricCollector returns a new SystemMetric collector. You would
// use this collector simply by instantiating the collector and the
// go smc.Collect() will handle the rest
func NewSystemMetricCollector(filter string, collectPeriodms int) SystemMetric {
	smc := SystemMetric{
		memStats: runtime.MemStats{},
		config:   collect.NewConfig(filter, collectPeriodms),
	}

	for _, v := range viewList {
		if err := view.Register(v); err != nil {
			panic(err)
		}
	}

	go smc.collect()
	return smc
}

// Collect records measurements for all the configured
// system metrics.
func (smc *SystemMetric) collect() {
	for {
		ctx := context.Background()
		runtime.ReadMemStats(&smc.memStats)

		// you should define a map of measures to function to grab
		// the measurement
		collectMap := map[*stats.Int64Measure]uint64{
			alloc: smc.getAlloc(),
		}

		for _, metric := range statsList {
			stats.Record(ctx, metric.M(int64(collectMap[metric])))
		}

		time.Sleep(time.Duration(smc.config.CollectPeriodms) * time.Minute)
	}

}

// actual function that returns the allocated bytes
func (smc *SystemMetric) getAlloc() uint64 {
	return smc.memStats.Alloc
}
