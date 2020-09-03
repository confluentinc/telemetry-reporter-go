package export

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	got := NewConfig(dummyIncludeFilter, dummyReportingPeriod)

	if config != got {
		t.Errorf("New Config failed, expected %v, got %v", config, got)
	}
}
