package export

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gogo/protobuf/proto"
	"go.opencensus.io/metric/metricdata"
)

// HTTP is an exporter that exports metrics to an
// HTTP endpoint
type HTTP struct {
	address   string
	aPIkey    string
	aPISecret string
	headerMap map[string]string
	config    *Config
}

// NewHTTP returns a new HTTP exporter
func NewHTTP(address string, apikey string, apisecret string, config *Config) HTTP {
	return HTTP{
		address:   address,
		aPIkey:    apikey,
		aPISecret: apisecret,
		headerMap: map[string]string{
			"Content-Type": "application/x-protobuf",
		},
		config: config,
	}
}

// NewHTTPWithHeaders returns a new HTTP exporter with the corresponding headers to be used
func NewHTTPWithHeaders(address string, apikey string, apisecret string, headers map[string]string, config *Config) HTTP {
	headers["Content-Type"] = "application/x-protobuf"

	return HTTP{
		address:   address,
		aPIkey:    apikey,
		aPISecret: apisecret,
		headerMap: headers,
		config:    config,
	}
}

// ExportMetrics converts the metrics to a metrics service request protobuf and
// makes a POST request with that payload to an HTTP endpoint.
func (e HTTP) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	includeData := []*metricdata.Metric{}

	for _, d := range data {
		if matched, _ := regexp.Match(e.config.IncludeFilter, []byte(d.Descriptor.Name)); matched {
			includeData = append(includeData, d)
		}
	}

	metricsRequestpb := metricsToServiceRequest(includeData)
	payload, err := proto.Marshal(metricsRequestpb)
	if err != nil {
		log.Fatal("Marshalling error: ", err)
	}

	e.postMetrics(payload)

	return nil
}

func (e HTTP) postMetrics(payload []byte) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", e.address, bytes.NewBuffer(payload))
	req.SetBasicAuth(e.aPIkey, e.aPISecret)

	for headerKey, headerVal := range e.headerMap {
		req.Header.Add(headerKey, headerVal)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)
	defer resp.Body.Close()
}
