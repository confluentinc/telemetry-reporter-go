package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/confluentinc/telemetry-reporter-go/export"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	metric1 = stats.Float64("metric1", "first dummy metric for test script for HTTP exporter", "ms")
	metric2 = stats.Int64("metric2", "second dummy metric for test script for HTTP exporter", "By")
)

// var (
// 	// KeyMethod ...
// 	KeyMethod, _ = tag.NewKey("method")

// 	// KeyStatus ...
// 	KeyStatus, _ = tag.NewKey("status")

// 	// KeyError ...
// 	KeyError, _ = tag.NewKey("error")
// )

var (
	metric1View = &view.View{
		Name:        "view1",
		Measure:     metric1,
		Description: "View for dummy metric1",
		Aggregation: view.LastValue(),
		// TagKeys:     []tag.Key{KeyMethod}}
	}

	metric2View = &view.View{
		Name:        "view2",
		Measure:     metric2,
		Description: "View for dummy metric2",
		Aggregation: view.Count(),
	}
)

// configure where to send metrics to
const (
	address   = "<address>"
	apikey    = "<api_key>"
	apisecret = "<api_secret>"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	view.Register(metric1View, metric2View)

	config := export.NewConfigWithReportingPeriod(`.*`, 10000)
	http := export.NewHTTP(address, apikey, apisecret, config)

	exporter := export.NewExporterAgent(http, config)
	exporter.Start()
	defer exporter.Stop()

	ctx := context.Background()

	for {
		randomFloat := rand.Float64() * 10
		randomInt := rand.Int63n(100)
		fmt.Printf("%f\n", randomFloat)
		fmt.Printf("%d\n", randomInt)
		stats.Record(ctx, metric1.M(randomFloat), metric2.M(randomInt))
		time.Sleep(time.Second * 1)
	}
}
