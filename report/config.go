package report

import (
	"github.com/confluentinc/telemetry-reporter-go/collect"
	"github.com/confluentinc/telemetry-reporter-go/export"
	"go.opencensus.io/metric/metricexport"
)

// Config defines the data format of the
// configurations of a reporter.
type Config struct {
	CollectingPeriodms int
	ReportingPeriodms  int
	Collectors         []collect.Collector
	Exporters          []metricexport.Exporter
	CollectorConfigs   []*collect.Config
	ExporterConfigs    []*export.Config
}

// NewConfig returns a pointer to a new reporter Config.
func NewConfig(
	reportingPeriodms int,
	collectors []collect.Collector,
	exporters []metricexport.Exporter,
	collectorConfigs []*collect.Config,
	exporterConfigs []*export.Config,
) *Config {
	return &Config{
		ReportingPeriodms: reportingPeriodms,
		Collectors:        collectors,
		Exporters:         exporters,
		CollectorConfigs:  collectorConfigs,
		ExporterConfigs:   exporterConfigs,
	}
}
