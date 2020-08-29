package export

import (
	"context"
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

var (
	zookeeperImage = "confluentinc/cp-zookeeper:5.5.1"
	zookeeperPort  = "2181"
	kafkaImage     = "confluentinc/cp-server:5.5.1"
	kafkaPort      = "9092"
)

func TestNewKafka(t *testing.T) {
	got, err := NewKafka(config, kafkaConfig, topicInfo)
	if err != nil {
		t.Fatalf("Error creating new Kafka")
	}
	got.Stop()

	gotKafka := got.Exporter.(Kafka)
	compareKafka(t, *kafkaExporter, gotKafka)
}

func TestSetMessageFlushTime(t *testing.T) {
	got, err := NewKafka(config, kafkaConfig, topicInfo)
	defer got.Stop()
	if err != nil {
		t.Errorf("Error creating new Kafka")
	}
	got.SetMessageFlushTime(newFlushTime)

	gotKafka := got.Exporter.(Kafka)
	compareKafka(t, *kafkaExporterFlushTime, gotKafka)
}

func TestKafkaCreateTopic(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	network, err := cli.NetworkCreate(context.Background(), "kafka_export_metrics_network", types.NetworkCreate{})
	if err != nil {
		t.Errorf("Failed to Create Docker network: %v", err)
	}

	defer func() {
		if err = cli.NetworkRemove(context.Background(), network.ID); err != nil {
			panic(err)
		}
	}()

	req := testcontainers.ContainerRequest{
		Name:         "zookeeper-server",
		Image:        zookeeperImage,
		ExposedPorts: []string{zookeeperPort},
		WaitingFor:   wait.ForListeningPort(nat.Port(zookeeperPort)),
		Networks:     []string{"kafka_export_metrics_network"},
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": zookeeperPort,
			"ALLOW_ANONYMOUS_LOGIN": "yes",
		},
	}

	zookeeper, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't start zookeeper image: %v", err)
	}

	defer zookeeper.Terminate(context.Background())

	req = testcontainers.ContainerRequest{
		Image:        kafkaImage,
		ExposedPorts: []string{"127.0.0.1:" + kafkaPort + ":" + kafkaPort},
		WaitingFor:   wait.ForListeningPort(nat.Port(kafkaPort)),
		Networks:     []string{"kafka_export_metrics_network"},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                                  "1",
			"KAFKA_ADVERTISED_LISTENERS":                       "PLAINTEXT_HOST://localhost:" + kafkaPort,
			"KAFKA_ZOOKEEPER_CONNECT":                          "zookeeper-server:2181",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":             "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":                 "PLAINTEXT_HOST",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                  "false",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":           "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":              "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR":   "1",
			"KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR": "1",
			"CONFLUENT_METRICS_ENABLE":                         "false",
		},
	}

	kafkaServer, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't start kafka image: %v", err)
	}
	defer kafkaServer.Terminate(context.Background())

	port, err := kafkaServer.PortEndpoint(context.Background(), nat.Port(kafkaPort), "")
	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't get kafka server's port")
	}

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": port,
	}

	topicInfo := TopicConfig{
		Topic:         "test",
		NumReplicas:   1,
		NumPartitions: 1,
	}

	if err := createTopic(topicInfo, kafkaConfig); err != nil {
		t.Errorf("Couldn't create topic: %v", err)
	}
}

func TestKafkaExportMetrics(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	network, err := cli.NetworkCreate(context.Background(), "kafka_export_metrics_network", types.NetworkCreate{})
	if err != nil {
		t.Errorf("Failed to Create Docker network: %v", err)
	}

	defer func() {
		if err = cli.NetworkRemove(context.Background(), network.ID); err != nil {
			panic(err)
		}
	}()

	req := testcontainers.ContainerRequest{
		Name:         "zookeeper-server",
		Image:        zookeeperImage,
		ExposedPorts: []string{zookeeperPort},
		WaitingFor:   wait.ForListeningPort(nat.Port(zookeeperPort)),
		Networks:     []string{"kafka_export_metrics_network"},
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": zookeeperPort,
			"ALLOW_ANONYMOUS_LOGIN": "yes",
		},
	}

	zookeeper, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't start zookeeper image: %v", err)
	}

	defer zookeeper.Terminate(context.Background())

	req = testcontainers.ContainerRequest{
		Image:        kafkaImage,
		ExposedPorts: []string{"127.0.0.1:" + kafkaPort + ":" + kafkaPort},
		WaitingFor:   wait.ForListeningPort(nat.Port(kafkaPort)),
		Networks:     []string{"kafka_export_metrics_network"},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                                  "1",
			"KAFKA_ADVERTISED_LISTENERS":                       "PLAINTEXT_HOST://localhost:" + kafkaPort,
			"KAFKA_ZOOKEEPER_CONNECT":                          "zookeeper-server:2181",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":             "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":                 "PLAINTEXT_HOST",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                  "true",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":           "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":              "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR":   "1",
			"KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR": "1",
			"CONFLUENT_METRICS_ENABLE":                         "false",
		},
	}

	kafkaServer, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't start kafka image: %v", err)
	}
	defer kafkaServer.Terminate(context.Background())

	port, err := kafkaServer.PortEndpoint(context.Background(), nat.Port(kafkaPort), "")
	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't get kafka server's port")
	}

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": port,
	}

	topicInfo := TopicConfig{
		Topic: "test",
	}

	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't create producer: %v", err)
	}

	exportKafka := Kafka{
		config:    config,
		producer:  producer,
		topicInfo: topicInfo,
	}

	err = exportKafka.ExportMetrics(context.Background(), metrics)
	if err != nil {
		t.Errorf("Kafka Export Metrics Failed, couldn't export metrics: %v", err)
	}
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
