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

	topicCreationInfo = TopicConfig{
		Topic:         topicName,
		NumReplicas:   1,
		NumPartitions: 1,
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
		t.Errorf("Error creating new Kafka: %v", err)
	}
	got.Stop()

	gotKafka := got.Exporter.(Kafka)
	compareKafka(t, *kafkaExporter, gotKafka)
}

func TestSetMessageFlushTime(t *testing.T) {
	got, err := NewKafka(config, kafkaConfig, topicInfo)
	defer got.Stop()
	if err != nil {
		t.Errorf("Error creating new Kafka: %v", err)
	}
	got.SetMessageFlushTime(newFlushTime)

	gotKafka := got.Exporter.(Kafka)
	compareKafka(t, *kafkaExporterFlushTime, gotKafka)
}

func TestKafkaCreateTopic(t *testing.T) {
	cli, network := createDockerNetwork(t, "kafka_export_metrics_network")
	defer removeDockerNetwork(t, cli, network)

	zookeeper := startZookeeperContainer(t, zookeeperImage, "zookeeper-server", "kafka_export_metrics_network", zookeeperPort)
	defer zookeeper.Terminate(context.Background())

	kafkaServer := startKafkaContainer(t, kafkaImage, "kafka_export_metrics_network", "zookeeper-server", zookeeperPort, "false", kafkaPort)
	defer kafkaServer.Terminate(context.Background())

	kafkaConfig := getKafkaConfig(t, kafkaServer, kafkaPort)

	if err := createTopic(topicCreationInfo, kafkaConfig); err != nil {
		t.Errorf("Failed to create topic: %v", err)
	}
}

func TestKafkaExportMetrics(t *testing.T) {
	cli, network := createDockerNetwork(t, "kafka_export_metrics_network")
	defer removeDockerNetwork(t, cli, network)

	zookeeper := startZookeeperContainer(t, zookeeperImage, "zookeeper-server", "kafka_export_metrics_network", zookeeperPort)
	defer zookeeper.Terminate(context.Background())

	kafkaServer := startKafkaContainer(t, kafkaImage, "kafka_export_metrics_network", "zookeeper-server", zookeeperPort, "true", kafkaPort)
	defer kafkaServer.Terminate(context.Background())

	kafkaConfig := getKafkaConfig(t, kafkaServer, kafkaPort)

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

func createDockerNetwork(t *testing.T, networkName string) (*client.Client, types.NetworkCreateResponse) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Errorf("Failed to create Docker client: %v", err)
	}

	network, err := cli.NetworkCreate(context.Background(), networkName, types.NetworkCreate{})
	if err != nil {
		t.Errorf("Failed to create Docker network: %v", err)
	}

	return cli, network
}

func removeDockerNetwork(t *testing.T, cli *client.Client, network types.NetworkCreateResponse) {
	if err := cli.NetworkRemove(context.Background(), network.ID); err != nil {
		t.Errorf("Failed to remove Docker network: %v", err)
	}
}

func startZookeeperContainer(t *testing.T, image string, containerName string, network string, port string) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Name:         containerName,
		Image:        image,
		ExposedPorts: []string{port},
		WaitingFor:   wait.ForListeningPort(nat.Port(port)),
		Networks:     []string{network},
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": port,
			"ALLOW_ANONYMOUS_LOGIN": "yes",
		},
	}

	zookeeper, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Errorf("Failed to start zookeeper container: %v", err)
	}

	return zookeeper
}

func startKafkaContainer(
	t *testing.T,
	image string,
	network string,
	zookeeperContainerName string,
	zookeeperPort string,
	kafkaAutoCreateTopic string,
	port string,
) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"127.0.0.1:" + port + ":" + port},
		WaitingFor:   wait.ForListeningPort(nat.Port(port)),
		Networks:     []string{network},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                                  "1",
			"KAFKA_ADVERTISED_LISTENERS":                       "PLAINTEXT_HOST://localhost:" + port,
			"KAFKA_ZOOKEEPER_CONNECT":                          zookeeperContainerName + ":" + zookeeperPort,
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":             "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":                 "PLAINTEXT_HOST",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                  kafkaAutoCreateTopic,
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
		t.Errorf("Failed to start kafka container: %v", err)
	}

	return kafkaServer
}

func getKafkaConfig(t *testing.T, kafkaServer testcontainers.Container, kafkaPort string) *kafka.ConfigMap {
	port, err := kafkaServer.PortEndpoint(context.Background(), nat.Port(kafkaPort), "")
	if err != nil {
		t.Errorf("Failed to get kafka server container's port")
	}

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": port,
	}

	return kafkaConfig
}

func compareKafka(t *testing.T, want Kafka, got Kafka) {
	if want.config != got.config {
		t.Errorf("New Kafka failed, expected config %v, got %v", want.config, got.config)
	}

	if eq := reflect.DeepEqual(*want.kafkaConfig, *got.kafkaConfig); !eq {
		t.Errorf("New Kafka failed, expected kafka config %v, got %v", want.kafkaConfig, got.kafkaConfig)
	}

	if want.topicInfo.Topic != got.topicInfo.Topic {
		t.Errorf("New Kafka failed, expected topic %v, got %v", want.topicInfo.Topic, got.topicInfo.Topic)
	}

	if want.messageFlushTimeSec != got.messageFlushTimeSec {
		t.Errorf("New Kafka failed, expected messageFlushTime %v, got %v", want.messageFlushTimeSec, got.messageFlushTimeSec)
	}
}
