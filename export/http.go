package export

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"regexp"

	"go.opencensus.io/metric/metricdata"
	"google.golang.org/protobuf/proto"
)

// HTTP is an exporter that exports metrics to an
// HTTP endpoint
type HTTP struct {
	address   string
	apiKey    string
	apiSecret string
	headerMap map[string]string
	client    *http.Client
	config    Config
}

// NewHTTP returns a new exporter agent with an HTTP exporter attached
func NewHTTP(address string, apiKey string, apiSecret string, config Config) *ExporterAgent {
	headerMap := map[string]string{
		"Content-Type": "application/x-protobuf",
	}

	exporter := HTTP{
		address:   address,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		headerMap: headerMap,
		client:    &http.Client{},
		config:    config,
	}

	agent := newExporterAgent(exporter)
	if err := agent.Start(exporter.config.reportingPeriodMilliseconds); err != nil {
		panic(err)
	}

	return agent
}

// AddHeader adds a map of headers to the exporter for its HTTP request.
func (e *ExporterAgent) AddHeader(headerMap map[string]string) {
	for k, v := range headerMap {
		e.Exporter.(HTTP).headerMap[k] = v
	}
}

// ExportMetrics converts the metrics to a metrics service request protobuf and
// makes a POST request with that payload to an HTTP endpoint.
func (e HTTP) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	includeData := []*metricdata.Metric{}
	resource, _ := TotDetector(ctx)

	for _, d := range data {
		if matched, _ := regexp.Match(e.config.IncludeFilter, []byte(d.Descriptor.Name)); matched {
			d.Resource = resource
			includeData = append(includeData, d)
		}
	}

	metricsRequestProto := metricsToServiceRequest(includeData)
	payload, err := proto.Marshal(metricsRequestProto)
	if err != nil {
		log.Fatal("Marshalling error: ", err)
	}

	e.postMetrics(payload)

	return nil
}

func (e HTTP) postMetrics(payload []byte) {
	req, _ := http.NewRequest("POST", e.address, bytes.NewBuffer(payload))
	req.SetBasicAuth(e.apiKey, e.apiSecret)

	for headerKey, headerVal := range e.headerMap {
		req.Header.Add(headerKey, headerVal)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		panic(err)
	}

	log.Println(resp)
	defer resp.Body.Close()
}
