package export

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
)

var (
	address   = "address"
	apiKey    = "key"
	apiSecret = "secret"
	headerMap = map[string]string{
		"Content-Type": "application/x-protobuf",
		"key":          "val",
	}
	exportPort = ":8081"

	dummyHTTP = HTTP{
		address:   address,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		headerMap: headerMap,
		client:    &http.Client{},
		config:    config,
	}
)

func TestNewHTTPAddHeader(t *testing.T) {
	got, err := NewHTTP(address, apiKey, apiSecret, config)
	if err != nil {
		t.Errorf("Error creating NewHTTP")
	}

	got.AddHeader(map[string]string{"key": "val"})
	got.Stop()

	compareHTTP(t, dummyHTTP, got.Exporter.(HTTP))
}

func TestHTTPExportMetrics(t *testing.T) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		got, _ := ioutil.ReadAll(r.Body)
		metricsRequest, err := metricsToServiceRequest(metrics)
		if err != nil {
			t.Errorf("Error exporting metrics: %v", err)
		}

		want, err := proto.Marshal(metricsRequest)
		if err != nil {
			t.Errorf("Marshalling error: %v", err)
		}

		if res := bytes.Compare(got, want); res != 0 {
			t.Errorf("Metrics export failed, expected %v, got %v", want, got)
		}
	})

	go func() {
		log.Fatal(http.ListenAndServe(exportPort, nil))
	}()

	exportHTTP := HTTP{
		address: "http://localhost" + exportPort,
		client:  &http.Client{},
		config:  config,
	}

	if err := exportHTTP.ExportMetrics(context.Background(), metrics); err != nil {
		t.Errorf("Error Exporting Metrics to HTTP: %e", err)
	}
}

func compareHTTP(t *testing.T, want HTTP, got HTTP) {
	if want.address != got.address {
		t.Errorf("New HTTP failed, expected address %v, got %v", want.address, got.address)
	}

	if want.apiKey != got.apiKey {
		t.Errorf("New HTTP failed, expected key %v, got %v", want.apiKey, got.apiKey)
	}

	if want.apiSecret != got.apiSecret {
		t.Errorf("New HTTP failed, expected secret %v, got %v", want.apiSecret, got.apiSecret)
	}

	if !reflect.DeepEqual(want.headerMap, got.headerMap) {
		t.Errorf("New HTTP failed, expected map %v, got %v", want.headerMap, got.headerMap)
	}

	if !reflect.DeepEqual(want.client, got.client) {
		t.Errorf("New HTTP failed, expected client %v, got %v", *want.client, *got.client)
	}

	if want.config != got.config {
		t.Errorf("New HTTP failed, expected config %v, got %v", want.config, got.config)
	}
}
