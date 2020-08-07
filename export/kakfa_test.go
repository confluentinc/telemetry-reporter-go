package export

import (
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	topicName = "test"

	kafkaConfig = &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}

	topicInfo = TopicConfig{
		Topic: topicName,
	}

	kafkaExporter = &Kafka{
		config:      config,
		kafkaConfig: kafkaConfig,
		topicInfo:   topicInfo,
	}
)

func TestNewKafka(t *testing.T) {
	got := NewKafka(config, kafkaConfig, topicInfo)
	defer got.Stop()

	gotKafka := got.Exporter.(Kafka)

	if kafkaExporter.config != gotKafka.config {
		t.Fatalf("New Kafka failed, expected config %v, got %v", kafkaExporter.config, gotKafka.config)
	}

	if eq := reflect.DeepEqual(*kafkaExporter.kafkaConfig, *gotKafka.kafkaConfig); !eq {
		t.Fatalf("New Kafka failed, expected kafka config %v, got %v", kafkaExporter.kafkaConfig, gotKafka.kafkaConfig)
	}

	if kafkaExporter.topicInfo.Topic != gotKafka.topicInfo.Topic {
		t.Fatalf("New Kafka failed, expected topic %v, got %v", kafkaExporter.topicInfo.Topic, gotKafka.topicInfo.Topic)
	}

}
