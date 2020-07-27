package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/confluentinc/telemetry-reporter-go/report"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/telemetry-reporter-go/export"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	metric1 = stats.Float64("metric1", "first dummy metric for test script for reporter", "ms")
	metric2 = stats.Int64("metric2", "second dummy metric for test script for reporter", "By")
)

var (
	tag1, _ = tag.NewKey("key")
)

var (
	metric1View = &view.View{
		Name:        "view1",
		Measure:     metric1,
		Description: "View for dummy metric1",
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{tag1},
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

	topic           = "test"
	reportingPeriod = 10000
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}

	exporterConfig := export.NewConfig(`.*`)
	httpExporter := export.NewHTTP(address, apikey, apisecret, exporterConfig)
	kafkaExporter := export.NewKafka(exporterConfig, kafkaConfig, topic)
	reporterConfig := report.NewConfigOnlyExporters(reportingPeriod, []*export.Config{exporterConfig}, httpExporter, kafkaExporter)
	reporter := report.NewReporter(reporterConfig, metric1View, metric2View)
	defer reporter.Stop()
	defer kafkaExporter.Stop()

	ctx, err := tag.New(context.Background(), tag.Insert(tag1, "val"))
	if err != nil {
		log.Fatal("Error creating tag: ", err)
	}

	for {
		randomFloat := rand.Float64() * 10
		randomInt := rand.Int63n(100)
		fmt.Printf("%f\n", randomFloat)
		fmt.Printf("%d\n", randomInt)
		stats.Record(ctx, metric1.M(randomFloat), metric2.M(randomInt))
		time.Sleep(time.Second * 1)
	}
}
