package export

import (
	"context"
	"log"

	"go.opencensus.io/metric/metricdata"
)

// Stdout ...
type Stdout struct {
	config *Config
}

// NewStdout ...
func NewStdout(config *Config) Stdout {
	return Stdout{
		config: config,
	}
}

// ExportMetrics ...
func (e Stdout) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	for _, d := range data {
		log.Printf(d.Descriptor.Name)
		log.Printf(d.Descriptor.Description)
		for _, ts := range d.TimeSeries {
			for _, point := range ts.Points {
				log.Printf("value=%v", point.Value)
			}
		}
		log.Printf("\n\n")
	}

	return nil
}
