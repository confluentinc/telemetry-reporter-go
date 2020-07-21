package report

import (
	"time"

	"github.com/confluentinc/telemetry-reporter-go/collect"
	"github.com/confluentinc/telemetry-reporter-go/export"
	"go.opencensus.io/stats/view"
)

// Reporter ...
type Reporter struct {
	collectors []*collect.CollectorAgent
	exporters  []*export.ExporterAgent
	views      []*view.View
	config     *Config
}

func (r *Reporter) initExporters() {
	r.exporters = []*export.ExporterAgent{}

	for i, metricExporter := range r.config.Exporters {
		// double check this logic later, if ReportingPeriodms isn't set, set it
		if r.config.ExporterConfigs[i].ReportingPeriodms == 0 {
			r.config.ExporterConfigs[i].ReportingPeriodms = r.config.ReportingPeriodms
		}

		exporter := export.NewExporterAgent(metricExporter, r.config.ExporterConfigs[i])
		exporter.Start()

		r.exporters = append(r.exporters, exporter)
	}
}

func (r *Reporter) initCollecters() {
	r.collectors = []*collect.CollectorAgent{}

	for i, metricCollector := range r.config.Collectors {
		collector := collect.NewCollectorAgent(metricCollector, r.config.CollectorConfigs[i])

		r.collectors = append(r.collectors, collector)
	}
}

func (r *Reporter) startCollect() {
	for {
		for _, collectorAgent := range r.collectors {
			collectorAgent.Collect()
		}
		time.Sleep(time.Duration(r.config.CollectingPeriodms) * time.Millisecond)
	}
}

func (r *Reporter) registerViews() {
	for _, v := range r.views {
		view.Register(v)
	}
}

// NewReporter ...
func NewReporter(vs []*view.View, config *Config) {
	reporter := &Reporter{
		config: config,
		views:  vs,
	}

	reporter.registerViews()
	reporter.initCollecters()
	reporter.initExporters()
	go reporter.startCollect()
}