# Telemetry-Reporter-Go

This library offers a collect and export package that define custom collectors that collect 
metrics using OpenCensus and exporters that export metrics in OpenCensus protobuf format.

## Installation

```bash
go get -u github.com/confluentinc/telemetry-reporter-go
```

## Export
For the export package we provide 2 custom exporters: HTTP Exporter and Kafka exporter.

For both exporters you must define an `export.Config` which you can do using the `export.NewConfig` function. The Config is currently made up of an include filter (regex filter for what metrics to export) and a reporting period in milliseconds.

### Kafka

The Kafka exporter needs an `export.Config`, KafkaConfig, and a `export.TopicInfo`. KafkaConfig is from the [Confluent-Kafka-Go Library](https://github.com/confluentinc/confluent-kafka-go) and a list of configurations can be found [here](https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md).

`export.TopicInfo` is a struct that keeps track of the name of the topic to export to as well as number of partitions and replicas to configure if the topic does not exist in the configured broker you are exporting to.

The Kafka exporter is instantiated by calling `export.NewKafka`. Here is an example:
```go
kafkaExporter := export.NewKafka(config, kafkaConfig, topicInfo)
defer kafkaExporter.Stop()
```

### HTTP

The HTTP exporter needs an address, API key, API secret, headers (if necessary), and an `export.Config`. It is instantiated by calling `export.NewHTTP`. Here is an example:

```go
http := export.NewHTTP(address, apikey, apisecret, config)
defer http.Stop()
```

Once an exporter is instantiated and metrics are instrumented with [OpenCensus](https://github.com/census-instrumentation/opencensus-go), you're all ready to go!
