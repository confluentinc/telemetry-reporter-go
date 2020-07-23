package export

import (
	"context"
	"fmt"
	"regexp"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gogo/protobuf/proto"
	"go.opencensus.io/metric/metricdata"
)

// Kafka is an exporter that exports metrics to a
// Kafka broker.
type Kafka struct {
	config      *Config
	kafkaConfig *kafka.ConfigMap
	topic       string
	// partition   int32
}

// NewKafka returns a new Kafka exporter
func NewKafka(config *Config, kafkaConfig *kafka.ConfigMap, topic string) *Kafka {
	return &Kafka{
		config:      config,
		kafkaConfig: kafkaConfig,
		topic:       topic,
	}
}

// ExportMetrics converts the metrics to a metrics service request protobuf and
// makes a POST request with that payload to a Kafka broker.
func (e Kafka) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	producer, err := kafka.NewProducer(e.kafkaConfig)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	for _, d := range data {
		if matched, _ := regexp.Match(e.config.IncludeFilter, []byte(d.Descriptor.Name)); matched {
			metricsRequestpb := metricToProto(d)
			payload, err := proto.Marshal(metricsRequestpb)
			if err != nil {
				panic(err)
			}

			producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &e.topic,
					Partition: kafka.PartitionAny, // or e.partition?
				},
				Value: payload,
			}, nil)
		}
	}

	producer.Flush(15 * 1000)
	return nil
}
