package collect

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	cpuPercent = stats.Float64("cpuPercent", "CPU % being utilized", "1")

	statsList = []*stats.Float64Measure{
		cpuPercent,
	}
)

var (
	cpuPercentView = &view.View{
		Name:        "cpuPercentView",
		Measure:     cpuPercent,
		Description: "view for cpuPercent",
		Aggregation: view.LastValue(),
	}

	viewList = []*view.View{
		cpuPercentView,
	}
)

var collectMap = map[*stats.Float64Measure]float64{
	cpuPercent: getCPUPercent(),
}

// SystemMetric is a collector that collects specific
// system metrics such as CPU utilization.
type SystemMetric struct {
	config *Config
}

// NewSystemMetricCollector returns a new SystemMetric collector.
func NewSystemMetricCollector() SystemMetric {
	return SystemMetric{
		config: &Config{
			StatsToCollect: statsList,
		},
	}
}

// Collect records measurements for all the configured
// system metrics.
func (smc *SystemMetric) Collect() {
	ctx := context.Background()

	for _, metric := range smc.config.StatsToCollect {
		stats.Record(ctx, metric.M(collectMap[metric]))
	}
}

func getCPUPercent() float64 {
	return 32.4
}
