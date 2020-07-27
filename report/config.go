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
	collectingPeriodms int,
	collectors []collect.Collector,
	exporters []metricexport.Exporter,
	collectorConfigs []*collect.Config,
	exporterConfigs []*export.Config,
) *Config {
	return &Config{
		ReportingPeriodms:  reportingPeriodms,
		CollectingPeriodms: collectingPeriodms,
		Collectors:         collectors,
		Exporters:          exporters,
		CollectorConfigs:   collectorConfigs,
		ExporterConfigs:    exporterConfigs,
	}
}

// NewConfigOnlyExporters returns a pointer to a new reporter
// Config that only uses exporters.
func NewConfigOnlyExporters(
	reportingPeriodms int,
	exporterConfigs []*export.Config,
	exporters ...metricexport.Exporter,
) *Config {
	config := &Config{
		ReportingPeriodms: reportingPeriodms,
		Exporters:         exporters,
	}

	if len(exporterConfigs) == 1 && len(exporters) > 1 {
		configs := make([]*export.Config, len(exporters))

		for i := range configs {
			configs[i] = exporterConfigs[0]
		}

		exporterConfigs = configs
	}

	config.ExporterConfigs = exporterConfigs
	return config

}
