// Package collect implements a library to create
// and define different pull based collectors.
package collect

// Collector is an interface that collectors implement
// to collect data.
type Collector interface {
	Collect()
}

// Config defines the data format of the general
// configurations of a collector.
type Config struct {
	IncludeFilter   string
	CollectPeriodms int
}

// NewConfig returns a pointer to a new collector Config.
func NewConfig(filter string, collectPeriodms int) Config {
	return Config{
		IncludeFilter:   filter,
		CollectPeriodms: collectPeriodms,
	}
}
