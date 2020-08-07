package export

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/resource"
	"google.golang.org/protobuf/proto"
)

// TopicConfig holds the configurations for Topic info
type TopicConfig struct {
	Topic         string
	NumPartitions int
	NumReplicas   int
}

// Kafka is an exporter that exports metrics to a
// Kafka broker.
type Kafka struct {
	config              Config
	kafkaConfig         *kafka.ConfigMap
	producer            *kafka.Producer
	topicInfo           TopicConfig
	messageFlushTimeSec int
	lastDroppedLogCount int
}

// NewKafka returns a new Kafka exporter
func NewKafka(config Config, kafkaConfig *kafka.ConfigMap, topicInfo TopicConfig) *ExporterAgent {
	createTopic(topicInfo, kafkaConfig)

	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		panic(err)
	}

	kafka := Kafka{
		config:              config,
		kafkaConfig:         kafkaConfig,
		topicInfo:           topicInfo,
		producer:            producer,
		lastDroppedLogCount: 0,
		messageFlushTimeSec: 15,
	}

	agent := newExporterAgent(kafka)
	if err := agent.Start(kafka.config.reportingPeriodMilliseconds); err != nil {
		panic(err)
	}

	return agent
}

func createTopic(topicInfo TopicConfig, kafkaConfig *kafka.ConfigMap) {
	adminClient, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		os.Exit(1)
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDuration, err := time.ParseDuration("60s")
	if err != nil {
		panic("time.ParseDuration(60s)")
	}

	if topicInfo.NumPartitions != 0 && topicInfo.NumReplicas != 0 {
		results, err := adminClient.CreateTopics(ctx,
			[]kafka.TopicSpecification{{
				Topic:             topicInfo.Topic,
				NumPartitions:     topicInfo.NumPartitions,
				ReplicationFactor: topicInfo.NumReplicas,
			}},
			kafka.SetAdminOperationTimeout(maxDuration))

		if err != nil {
			fmt.Printf("Problem during the topic creation: %v\n", err)
			os.Exit(1)
		}

		// Check for specific topic errors
		for _, result := range results {
			if result.Error.Code() != kafka.ErrNoError &&
				result.Error.Code() != kafka.ErrTopicAlreadyExists {
				fmt.Printf("Topic creation failed for %s: %v",
					result.Topic, result.Error.String())
				os.Exit(1)
			}
		}
	}

	adminClient.Close()

}

// Stop closes the Kafka producer.
func (e Kafka) Stop() {
	defer e.producer.Close()
}

// SetMessageFlushTime sets the time to wait to flush the
// Kafka message buffer. Default is 15 seconds
func (e Kafka) SetMessageFlushTime(seconds int) {
	e.messageFlushTimeSec = seconds
}

// ExportMetrics converts the metrics to a metrics service request protobuf and
// makes a POST request with that payload to a Kafka broker.
func (e Kafka) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	go handleEvents(e.producer.Events())

	for _, d := range data {
		if matched, _ := regexp.Match(e.config.IncludeFilter, []byte(d.Descriptor.Name)); matched {
			d.Resource, _ = resource.FromEnv(ctx)
			metricsRequestpb := metricToProto(d)
			payload, err := proto.Marshal(metricsRequestpb)
			if err != nil {
				log.Fatal("Marshalling Error: ", err)
			}

			err = e.producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &e.topicInfo.Topic,
					Partition: kafka.PartitionAny,
				},
				Value: payload,
			}, nil)

			if err != nil {
				log.Fatal("Error sending message with Producer: ", err)
			}
		}
	}

	droppedCount := e.producer.Flush(e.messageFlushTimeSec * 1000)
	droppedDelta := droppedCount - e.lastDroppedLogCount
	if droppedDelta > 0 {
		log.Println("Failed to produce %i metrics messages", droppedDelta)
	}

	e.lastDroppedLogCount = droppedCount

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
