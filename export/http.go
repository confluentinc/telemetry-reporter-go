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
func NewHTTP(address string, apikey string, apisecret string, headers map[string]string, config *Config) HTTP {
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
	for _, d := range data {
		if matched, _ := regexp.Match(e.config.IncludeFilter, []byte(d.Descriptor.Name)); matched {
			metricsRequestpb := metricsToServiceRequest(data)
			payload, err := proto.Marshal(metricsRequestpb)
			if err != nil {
				log.Fatal("Marshalling error: ", err)
			}

			client := &http.Client{}
			req, _ := http.NewRequest("POST", e.address, bytes.NewBuffer(payload))
			req.SetBasicAuth(e.aPIkey, e.aPISecret)
			for headerKey, headerVal := range e.headerMap {
				req.Header.Add(headerKey, headerVal)
			}
			// req.Header.Add("Content-Type", "application/x-protobuf")
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			fmt.Println(resp)
			defer resp.Body.Close()
		}
	}

	return nil
}
