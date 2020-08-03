# Telemetry-Reporter-Go

This library offers a collect and export package that define custom collectors that collect 
metrics using OpenCensus and exporters that export metrics in OpenCensus protobuf format.

## Installation

```bash
go get -u github.com/confluentinc/telemetry-reporter-go
```

## Export
For the export package we provide 2 custom exporters: HTTP Exporter and Kafka exporter.

### Kafka

### HTTP