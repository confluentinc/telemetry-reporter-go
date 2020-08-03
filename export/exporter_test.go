package export

import (
	"testing"
)

var (
	dummyIncludeFilter   = `.*`
	dummyReportingPeriod = 5

	config = Config{
		IncludeFilter:     dummyIncludeFilter,
		ReportingPeriodms: dummyReportingPeriod,
	}
)

func TestNewConfig(t *testing.T) {
	got := NewConfig(dummyIncludeFilter, dummyReportingPeriod)

	if config != got {
		t.Errorf("New Config failed, expected %v, got %v", config, got)
	}
}
