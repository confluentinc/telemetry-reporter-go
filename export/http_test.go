package export

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"

	"go.opencensus.io/metric/metricdata"
	"google.golang.org/protobuf/proto"
)

var (
	address   = ""
	apiKey    = ""
	apiSecret = ""
	headerMap = map[string]string{
		"Content-Type": "application/x-protobuf",
		"key":          "val",
	}

	dummyHTTP = HTTP{
		address:   address,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		headerMap: headerMap,
		client:    &http.Client{},
		config:    config,
	}

	metrics = []*metricdata.Metric{
		&metricdata.Metric{
			Descriptor: metricdata.Descriptor{
				Name:        dummyName,
				Description: dummyDesc,
				Unit:        metricdata.Unit(dummyUnit),
				Type:        metricdata.Type(dummyType),
				LabelKeys: []metricdata.LabelKey{
					metricdata.LabelKey{
						Key:         dummyLabelKey,
						Description: dummyKeyDesc,
					},
				},
			},
			TimeSeries: []*metricdata.TimeSeries{
				&metricdata.TimeSeries{
					LabelValues: []metricdata.LabelValue{
						metricdata.NewLabelValue(dummyLabelVal),
					},
					Points:    []metricdata.Point{metricdata.NewInt64Point(timeNow, intVal)},
					StartTime: timeNow,
				},
			},
		},
	}
)

func TestNewHTTP(t *testing.T) {
	got := NewHTTP(address, apiKey, apiSecret, map[string]string{"key": "val"}, config)
	defer got.Stop()

	compareHTTP(t, dummyHTTP, got.Exporter.(HTTP))
}

func TestExportMetrics(t *testing.T) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		got, _ := ioutil.ReadAll(r.Body)
		want, err := proto.Marshal(metricsToServiceRequest(metrics))
		if err != nil {
			log.Fatal("Marshalling error: ", err)
		}

		if res := bytes.Compare(got, want); res != 0 {
			t.Errorf("Metrics export failed, expected %v, got %v", want, got)
		}
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	exportHTTP := HTTP{
		address: "http://localhost:8081",
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

	if eq := reflect.DeepEqual(want.headerMap, got.headerMap); !eq {
		t.Errorf("New HTTP failed, expected map %v, got %v", want.headerMap, got.headerMap)
	}

	if eq := reflect.DeepEqual(want.client, got.client); !eq {
		t.Errorf("New HTTP failed, expected client %v, got %v", *want.client, *got.client)
	}

	if want.config != got.config {
		t.Errorf("New HTTP failed, expected config %v, got %v", want.config, got.config)
	}
}
