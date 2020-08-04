package collect

import (
	"context"
	"runtime"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	alloc = stats.Int64("alloc", "bytes of allocated heap objects", "by")

	statsList = []*stats.Int64Measure{
		alloc,
	}
)

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
	config   Config
}

// NewSystemMetricCollector returns a new SystemMetric collector.
func NewSystemMetricCollector(filter string, collectPeriodms int) SystemMetric {
	smc := SystemMetric{
		memStats: runtime.MemStats{},
		config:   NewConfig(filter, collectPeriodms),
	}

	for _, v := range viewList {
		if err := view.Register(v); err != nil {
			panic(err)
		}
	}

	go smc.Collect()
	return smc
}

// Collect records measurements for all the configured
// system metrics.
func (smc *SystemMetric) Collect() {
	for {
		ctx := context.Background()
		runtime.ReadMemStats(&smc.memStats)
		collectMap := map[*stats.Int64Measure]uint64{
			alloc: smc.getAlloc(),
		}

		for _, metric := range statsList {
			stats.Record(ctx, metric.M(int64(collectMap[metric])))
		}

		time.Sleep(time.Duration(smc.config.CollectPeriodms) * time.Minute)
	}

}

func (smc *SystemMetric) getAlloc() uint64 {
	return smc.memStats.Alloc
}
