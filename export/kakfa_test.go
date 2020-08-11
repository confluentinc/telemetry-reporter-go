package export

import (
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	topicName        = "test"
	newFlushTime     = 20
	defaultFlushTime = 15

	kafkaConfig = &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}

	topicInfo = TopicConfig{
		Topic: topicName,
	}

	kafkaExporter = &Kafka{
		config:              config,
		kafkaConfig:         kafkaConfig,
		topicInfo:           topicInfo,
		messageFlushTimeSec: defaultFlushTime,
	}

	kafkaExporterFlushTime = &Kafka{
		config:              config,
		kafkaConfig:         kafkaConfig,
		topicInfo:           topicInfo,
		messageFlushTimeSec: newFlushTime,
	}
)

func TestNewKafka(t *testing.T) {
	got := NewKafka(config, kafkaConfig, topicInfo)
	defer got.Stop()

	gotKafka := got.Exporter.(Kafka)
	compareKafka(t, *kafkaExporter, gotKafka)

}

func TestSetMessageFlushTime(t *testing.T) {
	got := NewKafka(config, kafkaConfig, topicInfo)
	defer got.Stop()
	got.SetMessageFlushTime(newFlushTime)

	gotKafka := got.Exporter.(Kafka)
	compareKafka(t, *kafkaExporterFlushTime, gotKafka)

}

func compareKafka(t *testing.T, want Kafka, got Kafka) {
	if want.config != got.config {
		t.Fatalf("New Kafka failed, expected config %v, got %v", want.config, got.config)
	}

	if eq := reflect.DeepEqual(*want.kafkaConfig, *got.kafkaConfig); !eq {
		t.Fatalf("New Kafka failed, expected kafka config %v, got %v", want.kafkaConfig, got.kafkaConfig)
	}

	if want.topicInfo.Topic != got.topicInfo.Topic {
		t.Fatalf("New Kafka failed, expected topic %v, got %v", want.topicInfo.Topic, got.topicInfo.Topic)
	}

	if want.messageFlushTimeSec != got.messageFlushTimeSec {
		t.Fatalf("New Kafka failed, expected messageFlushTime %v, got %v", want.messageFlushTimeSec, got.messageFlushTimeSec)
	}
}
