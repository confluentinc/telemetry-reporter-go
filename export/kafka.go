package export

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/protobuf/proto"
	"go.opencensus.io/metric/metricdata"
)

// Kafka is an exporter that exports metrics to a
// Kafka broker.
type Kafka struct {
	config      *Config
	kafkaConfig *kafka.ConfigMap
	topic       string
	producer    *kafka.Producer
	// partition   int32
}

// NewKafka returns a new Kafka exporter
func NewKafka(config *Config, kafkaConfig *kafka.ConfigMap, topic string) *Kafka {
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		panic(err)
	}

	return &Kafka{
		config:      config,
		kafkaConfig: kafkaConfig,
		topic:       topic,
		producer:    producer,
	}
}

// Stop closes the Kafka producer.
func (e *Kafka) Stop() {
	defer e.producer.Close()
}

// ExportMetrics converts the metrics to a metrics service request protobuf and
// makes a POST request with that payload to a Kafka broker.
func (e Kafka) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	go handleEvents(e.producer.Events())

	for _, d := range data {
		if matched, _ := regexp.Match(e.config.IncludeFilter, []byte(d.Descriptor.Name)); matched {
			metricsRequestpb := metricToProto(d)
			payload, err := proto.Marshal(metricsRequestpb)
			if err != nil {
				log.Fatal("Marshalling Error: ", err)
			}

			err = e.producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &e.topic,
					Partition: kafka.PartitionAny, // or e.partition?
				},
				Value: payload,
			}, nil)

			if err != nil {
				log.Fatal("Error sending message with Producer: ", err)
			}
		}
	}

	e.producer.Flush(15 * 1000)
	return nil
}

func handleEvents(events chan kafka.Event) {
	for e := range events {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
			} else {
				fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
			}
		}
	}
}
