package export

import (
	"sync"
	"time"

	"go.opencensus.io/metric/metricexport"
)

// ExporterAgent ...
type ExporterAgent struct {
	metricexport.Exporter
	ir             *metricexport.IntervalReader
	initReaderOnce sync.Once
	Config         *Config
}

// Config ...
type Config struct {
	IncludeFilter     string
	ReportingPeriodms int
}

// NewConfig ...
func NewConfig(filter string) *Config {
	return &Config{
		IncludeFilter: filter,
	}
}

// NewConfigWithReportingPeriod ...
func NewConfigWithReportingPeriod(filter string, reportingPeriod int) *Config {
	return &Config{
		IncludeFilter:     filter,
		ReportingPeriodms: reportingPeriod,
	}
}

// NewExporterAgent ...
func NewExporterAgent(e metricexport.Exporter, config *Config) *ExporterAgent {
	return &ExporterAgent{
		Exporter: e,
		Config:   config,
	}
}

// Start ...
func (e *ExporterAgent) Start() error {
	e.initReaderOnce.Do(func() {
		e.ir, _ = metricexport.NewIntervalReader(&metricexport.Reader{}, e.Exporter)
	})
	e.ir.ReportingInterval = time.Duration(e.Config.ReportingPeriodms) * time.Millisecond
	return e.ir.Start()
}

// Stop ...
func (e *ExporterAgent) Stop() {
	e.ir.Stop()
}
