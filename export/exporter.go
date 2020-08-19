// Package export implements a library to create
// and define different custom OpenCensus metrics
// exporters.
package export

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/metric/metricexport"
)

// ExporterAgent defines the wrapper format of exporters
// and data needed by all general exporters.
type ExporterAgent struct {
	metricexport.Exporter
	ir             *metricexport.IntervalReader
	initReaderOnce sync.Once
}

// Config defines the data format of the general
// configurations of an exporter.
type Config struct {
	IncludeFilter               string
	reportingPeriodMilliseconds int
}

// NewConfig returns a new exporter Config.
func NewConfig(filter string, reportingPeriod int) Config {
	return Config{
		IncludeFilter:               filter,
		reportingPeriodMilliseconds: reportingPeriod,
	}
}

// a user should never have to use this explicitly. They would
// simply instantiate an implemented exporter
func newExporterAgent(exporter metricexport.Exporter) *ExporterAgent {
	return &ExporterAgent{
		Exporter: exporter,
	}
}

// Start creates the ExporterAgent's IntervalReader (if needed),
// sets the reporting interval, and then starts the reader.
func (e *ExporterAgent) Start(reportingPeriodms int) error {
	e.initReaderOnce.Do(func() {
		e.ir, _ = metricexport.NewIntervalReader(&metricexport.Reader{}, e.Exporter)
	})

	if e.ir == nil {
		return errors.New("Failed to create Interval Reader")
	}

	e.ir.ReportingInterval = time.Duration(reportingPeriodms) * time.Millisecond
	return e.ir.Start()
}

// Stop stops the ExporterAgent's interval reader.
func (e *ExporterAgent) Stop() {
	if k, ok := e.Exporter.(Kafka); ok {
		k.Stop()
	}

	e.ir.Stop()
}
