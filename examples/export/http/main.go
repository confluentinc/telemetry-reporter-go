package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/confluentinc/telemetry-reporter-go/export"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	metric1 = stats.Float64("metric1", "first dummy metric for test script for HTTP exporter", "ms")
	metric2 = stats.Int64("metric2", "second dummy metric for test script for HTTP exporter", "By")
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
	address   = "<address>"    // address of the HTTP endpoint
	apikey    = "<api_key>"    // username for authentication
	apisecret = "<api_secret>" // password for authentication
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	view.Register(metric1View, metric2View)

	/*
		NewConfig is made up of and universal to all exporters
			1. regex include filter (`.*`)
			2. reporting period in ms (10000 -> 10s)
	*/
	config := export.NewConfig(`.*`, 10000)
	http, err := export.NewHTTP(address, apikey, apisecret, config)
	defer http.Stop()
	if err != nil {
		panic(err)
	}

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
