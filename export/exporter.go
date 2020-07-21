// Package export implements a library to create
// and define different custom OpenCensus metrics
// exporters.
package export

import (
	"sync"
	"time"

	"go.opencensus.io/metric/metricexport"
)

// ExporterAgent defines the wrapper format of exporters
// and data needed by all general exporters.
type ExporterAgent struct {
	metricexport.Exporter
	ir             *metricexport.IntervalReader
	initReaderOnce sync.Once
	Config         *Config
}

// Config defines the data format of the general
// configurations of an exporter.
type Config struct {
	IncludeFilter     string
	ReportingPeriodms int
}

// NewConfig returns a pointer to a new exporter Config.
func NewConfig(filter string) *Config {
	return &Config{
		IncludeFilter: filter,
	}
}

// NewConfigWithReportingPeriod returns a pointer to a new
// exporter Config with the specified reporting period.
func NewConfigWithReportingPeriod(filter string, reportingPeriod int) *Config {
	return &Config{
		IncludeFilter:     filter,
		ReportingPeriodms: reportingPeriod,
	}
}

// NewExporterAgent returns a pointer to a new ExporterAgent.
func NewExporterAgent(e metricexport.Exporter, config *Config) *ExporterAgent {
	return &ExporterAgent{
		Exporter: e,
		Config:   config,
	}
}

// Start creates the ExporterAgent's IntervalReader (if needed),
// sets the reporting interval, and then starts the reader.
func (e *ExporterAgent) Start() error {
	e.initReaderOnce.Do(func() {
		e.ir, _ = metricexport.NewIntervalReader(&metricexport.Reader{}, e.Exporter)
	})
	e.ir.ReportingInterval = time.Duration(e.Config.ReportingPeriodms) * time.Millisecond
	return e.ir.Start()
}

// Stop stops the ExporterAgent's interval reader.
func (e *ExporterAgent) Stop() {
	e.ir.Stop()
}
