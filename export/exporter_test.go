package export

import (
	"testing"
)

var (
	dummyIncludeFilter   = `.*`
	dummyReportingPeriod = 1000

	config = Config{
		IncludeFilter:               dummyIncludeFilter,
		reportingPeriodMilliseconds: dummyReportingPeriod,
	}
)

func TestNewConfig(t *testing.T) {
	got := NewConfig(dummyIncludeFilter, dummyReportingPeriod)

	if config != got {
		t.Errorf("New Config failed, expected %v, got %v", config, got)
	}
}
