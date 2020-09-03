package export

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"google.golang.org/protobuf/proto"
)

var (
	messagesSent    = stats.Int64("messages_sent", "the number of metric messages successfully sent", "1")
	messagesDropped = stats.Int64("messages_dropped", "the number of metric messages dropped", "1")
)

var (
	messagesSentView = &view.View{
		Name:        "messages_sent",
		Measure:     messagesSent,
		Description: "the number of metric messages successfully sent",
		Aggregation: view.Count(),
	}

	messagesDropedView = &view.View{
		Name:        "messages_dropped",
		Measure:     messagesDropped,
		Description: "the number of metric messages dropped",
		Aggregation: view.Sum(),
	}
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
	DroppedDelta        int
}

// NewKafka returns a new Kafka exporter
func NewKafka(config Config, kafkaConfig *kafka.ConfigMap, topicInfo TopicConfig) (*ExporterAgent, error) {
	if err := view.Register(messagesSentView, messagesDropedView); err != nil {
		return nil, errors.Wrap(err, "Error registering views")
	}

	// createTopic only actually creates the topic if it doesn't exist
	if err := createTopic(topicInfo, kafkaConfig); err != nil {
		return nil, errors.Wrap(err, "Error creating topic")
	}

	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating Kafka producer")
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
		return agent, errors.Wrap(err, "Error starting exporter")
	}

	return agent, nil
}

func createTopic(topicInfo TopicConfig, kafkaConfig *kafka.ConfigMap) error {
	adminClient, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return errors.Wrap(err, "Failed to create Admin client")
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDuration, err := time.ParseDuration("60s")
	if err != nil {
		return errors.Wrap(err, "time.ParseDuration(60s)")
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
			return errors.Wrap(err, "Problem during the topic creation")
		}

		// Check for specific topic errors
		for _, result := range results {
			if result.Error.Code() != kafka.ErrNoError &&
				result.Error.Code() != kafka.ErrTopicAlreadyExists {
				return errors.Wrap(result.Error, fmt.Sprintf("Topic creation failed for topic %v", result.Topic))
			}
		}
	}

	adminClient.Close()
	return nil
}

// Stop closes the Kafka producer.
func (e Kafka) Stop() {
	if e.producer == nil {
		return
	}

	e.producer.Close()
}

// SetMessageFlushTime sets the time to wait to flush the
// Kafka message buffer. Default is 15 seconds
func (e *ExporterAgent) SetMessageFlushTime(seconds int) {
	newKafka := e.Exporter.(Kafka)
	newKafka.messageFlushTimeSec = seconds
	e.Exporter = newKafka
}

// ExportMetrics converts the metrics to a metrics service request protobuf and
// makes a POST request with that payload to a Kafka broker.
func (e Kafka) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	go handleEvents(e.producer.Events())

	resource, err := TotDetector(ctx)
	if err != nil {
		return errors.Wrap(err, "Error creating resource detector")
	}

	for _, d := range data {
		matched, err := regexp.Match(e.config.IncludeFilter, []byte(d.Descriptor.Name))
		if err != nil {
			return errors.Wrap(err, "Error matching regular expression")
		}

		if matched {
			d.Resource = resource
			metricsRequestpb, err := metricToProto(d)
			if err != nil {
				return errors.Wrap(err, "Error converting metric to Proto")
			}

			payload, err := proto.Marshal(metricsRequestpb)
			if err != nil {
				return errors.Wrap(err, "Marshalling Error")
			}

			err = e.producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &e.topicInfo.Topic,
					Partition: kafka.PartitionAny,
				},
				Value: payload,
			}, nil)

			if err != nil {
				return errors.Wrap(err, "Error sending message with Producer")
			}
		}
	}

	droppedCount := e.producer.Flush(e.messageFlushTimeSec * 1000)
	stats.Record(context.Background(), messagesDropped.M(int64(droppedCount)))
	e.DroppedDelta = droppedCount - e.lastDroppedLogCount
	e.lastDroppedLogCount = droppedCount

	return nil
}

func handleEvents(events chan kafka.Event) {
	for e := range events {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				stats.Record(context.Background(), messagesDropped.M(1))
			} else {
				stats.Record(context.Background(), messagesSent.M(1))
			}
		}
	}
}
