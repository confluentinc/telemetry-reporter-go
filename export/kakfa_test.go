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

	if *kafkaExporter.config != *got.config {
		t.Fatalf("New Kafka failed, expected config %v, got %v", kafkaExporter.config, got.config)
	}

	if eq := reflect.DeepEqual(*kafkaExporter.kafkaConfig, *got.kafkaConfig); !eq {
		t.Fatalf("New Kafka failed, expected kafka config %v, got %v", kafkaExporter.kafkaConfig, got.kafkaConfig)
	}

	if kafkaExporter.topicInfo.Topic != got.topicInfo.Topic {
		t.Fatalf("New Kafka failed, expected topic %v, got %v", kafkaExporter.topicInfo.Topic, got.topicInfo.Topic)
	}
}
