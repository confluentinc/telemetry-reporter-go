package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/telemetry-reporter-go/export"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	metric1 = stats.Float64("metric1", "first dummy metric for test script for Kafka exporter", "ms")
	metric2 = stats.Int64("metric2", "second dummy metric for test script for Kafka exporter", "By")
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
		Aggregation: view.Sum(),
	}
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	view.Register(metric1View, metric2View)

	/*
		A full list of configurations can be found here
		https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md

		If you want to configure for Confluent Cloud you need
			1. "bootstrap.servers" -  Alias for metadata.broker.list:
			Initial list of brokers as a CSV list of broker host or host:port.
			2. "sasl.mechanism" - SASL mechanism to use for authentication, should be set to “PLAIN”
			3. "security.protocol" - Protocol used to communicate with brokers, should be set to “SASL_SSL”
			4. "sasl.username" - SASL username for use with the PLAIN and SASL-SCRAM-.. mechanisms
			5. "sasl.password" - SASL password for use with the PLAIN and SASL-SCRAM-.. mechanisms
	*/
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}

	config := export.NewConfig(`.*`, 10000)

	/*
		TopicConfig is made up of
			1. Topic - topic name to export messages to
			2. NumPartitions - sets the number of partitions of topic if
			TopicConfig.Topic doesn't exist and the exporter needs to create it
			3. NumReplicas - sets the number of replicas of topic if
			TopicConfig.Topic doesn't exist and the exporter needs to create it
	*/
	topicInfo := export.TopicConfig{
		Topic: "test",
	}
	kafkaExporter, err := export.NewKafka(config, kafkaConfig, topicInfo)
	defer kafkaExporter.Stop()
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
