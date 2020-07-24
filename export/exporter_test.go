package export

import (
	"testing"
)

var (
	dummyIncludeFilter   = `.*`
	dummyReportingPeriod = 5

	config = &Config{
		IncludeFilter: dummyIncludeFilter,
	}

	configWithRP = &Config{
		IncludeFilter:     dummyIncludeFilter,
		ReportingPeriodms: dummyReportingPeriod,
	}

	exporterAgent = &ExporterAgent{
		Exporter: NewStdout(config),
		Config:   config,
	}
)

func TestNewConfig(t *testing.T) {
	got := NewConfig(dummyIncludeFilter)

	if *config != *got {
		t.Errorf("New Config failed, expected %v, got %v", *config, *got)
	}
}

func TestNewConfigWithReportingPeriod(t *testing.T) {
	got := NewConfigWithReportingPeriod(dummyIncludeFilter, dummyReportingPeriod)

	if *configWithRP != *got {
		t.Errorf("New Config with Reporting Period failed, expected %v, got %v", *configWithRP, *got)
	}

}

func TestNewExporterAgent(t *testing.T) {
	got := NewExporterAgent(NewStdout(config), config)

	if *exporterAgent != *got {
		t.Errorf("New Exporter Agent failed, expected %v, got %v", *exporterAgent, *got)
	}
}
