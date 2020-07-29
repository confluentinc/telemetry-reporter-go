package export

import (
	"context"
	"log"

	"go.opencensus.io/metric/metricdata"
)

// Stdout is an exporter that exports metrics to stdout.
type Stdout struct {
	config *Config
}

// NewStdout returns a new Stdout exporter.
func NewStdout(config *Config) Stdout {
	return Stdout{
		config: config,
	}
}

// ExportMetrics prints the metrics' names, description, and values.
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
