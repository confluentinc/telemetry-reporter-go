package export

import (
	"testing"
	"time"

	"go.opencensus.io/metric/metricdata"

	a1 "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	v1 "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/golang/protobuf/ptypes"
)

var (
	timeNow       = time.Now()
	timestamp, _  = ptypes.TimestampProto(timeNow)
	intVal        = int64(10)
	doubleVal     = 12.345
	dummyLabelVal = "Val"
	dummyName     = "metric"
	dummyDesc     = "desc"
	dummyUnit     = "ms"
	dummyType     = 4
	dummyLabelKey = "Key"
	dummyKeyDesc  = "key desc"

	intPoints = []*v1.Point{
		&v1.Point{
			Timestamp: timestamp,
			Value: &v1.Point_Int64Value{
				Int64Value: intVal,
			},
		},
	}

	doublePoints = []*v1.Point{
		&v1.Point{
			Timestamp: timestamp,
			Value: &v1.Point_DoubleValue{
				DoubleValue: doubleVal,
			},
		},
	}

	labelVals = []*v1.LabelValue{
		&v1.LabelValue{
			Value:    dummyLabelVal,
			HasValue: true,
		},
	}

	timeseries = []*v1.TimeSeries{
		&v1.TimeSeries{
			StartTimestamp: timestamp,
			LabelValues:    labelVals,
			Points:         intPoints,
		},
	}

	labelKeys = []*v1.LabelKey{
		&v1.LabelKey{
			Key:         dummyLabelKey,
			Description: dummyKeyDesc,
		},
	}

	descriptor = &v1.MetricDescriptor{
		Name:        dummyName,
		Description: dummyDesc,
		Unit:        dummyUnit,
		Type:        v1.MetricDescriptor_Type(dummyType + 1),
		LabelKeys:   labelKeys,
	}

	dummyMetric = &v1.Metric{
		MetricDescriptor: descriptor,
		Timeseries:       timeseries,
	}

	dummyServiceRequest = &a1.ExportMetricsServiceRequest{
		Metrics: []*v1.Metric{
			dummyMetric,
		},
	}
)

func TestMetricToPointInt64(t *testing.T) {
	timeseries := metricdata.TimeSeries{
		Points: []metricdata.Point{metricdata.NewInt64Point(timeNow, intVal)},
	}
	got := metricToPoints(&timeseries)
	comparePoints(t, intPoints, got, true)
}

func TestMetricToPointDouble64(t *testing.T) {
	timeseries := metricdata.TimeSeries{
		Points: []metricdata.Point{metricdata.NewFloat64Point(timeNow, doubleVal)},
	}
	got := metricToPoints(&timeseries)
	comparePoints(t, doublePoints, got, false)
}

func TestMetricToLabelValues(t *testing.T) {
	timeseries := metricdata.TimeSeries{
		LabelValues: []metricdata.LabelValue{
			metricdata.NewLabelValue(dummyLabelVal),
		},
	}
	got := metricToLabelValues(&timeseries)
	compareLabelVals(t, labelVals, got)
}

func TestMetricToTimeSeries(t *testing.T) {
	metric := metricdata.Metric{
		TimeSeries: []*metricdata.TimeSeries{
			&metricdata.TimeSeries{
				LabelValues: []metricdata.LabelValue{
					metricdata.NewLabelValue(dummyLabelVal),
				},
				Points:    []metricdata.Point{metricdata.NewInt64Point(timeNow, intVal)},
				StartTime: timeNow,
			},
		},
	}

	got := metricToTimeSeries(&metric)
	compareMetricTimeseries(t, timeseries, got)
}

// TODO if we chose to provide support for this
func TestMetricToResource(t *testing.T) {
}

func TestMetricToDescriptor(t *testing.T) {
	metric := metricdata.Metric{
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
	}

	got := metricToDescriptor(&metric)
	compareMetricDesc(t, descriptor, got)
}

func TestMetricToLabelKeys(t *testing.T) {
	metric := metricdata.Metric{
		Descriptor: metricdata.Descriptor{
			LabelKeys: []metricdata.LabelKey{
				metricdata.LabelKey{
					Key:         dummyLabelKey,
					Description: dummyKeyDesc,
				},
			},
		},
	}
	got := metricToLabelKeys(&metric)
	compareLabelKeys(t, labelKeys, got)
}

func TestMetricToProto(t *testing.T) {
	metric := metricdata.Metric{
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
	}

	got := metricToProto(&metric)
	compareMetricTimeseries(t, dummyMetric.Timeseries, got.Timeseries)
	compareMetricDesc(t, dummyMetric.GetMetricDescriptor(), got.GetMetricDescriptor())
}

func TestMetricToServiceRequest(t *testing.T) {
	metrics := []*metricdata.Metric{
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
	got := metricsToServiceRequest(metrics)
	for i := range dummyServiceRequest.Metrics {
		compareMetricTimeseries(t, dummyServiceRequest.Metrics[i].Timeseries, got.Metrics[i].Timeseries)
		compareMetricDesc(t, dummyServiceRequest.Metrics[i].GetMetricDescriptor(), got.Metrics[i].GetMetricDescriptor())
	}
}

func comparePoints(t *testing.T, want []*v1.Point, got []*v1.Point, isInt bool) {
	if len(want) != len(got) {
		t.Errorf("Metric to Points int failed, expected length %v, got %v", len(want), len(got))
	}

	for i := range want {
		if want[i].Timestamp.String() != got[i].Timestamp.String() {
			t.Errorf("Metric to Points int failed, expected time %v, got %v", want[i].Timestamp, got[i].Timestamp)
		}
		if isInt {
			if want[i].GetInt64Value() != got[i].GetInt64Value() {
				t.Errorf("Metric to Points int failed, expected val %v, got %v", want[i].GetInt64Value(), got[i].GetInt64Value())
			}
		} else {
			if want[i].GetDoubleValue() != got[i].GetDoubleValue() {
				t.Errorf("Metric to Points double failed, expected val %v, got %v", want[i].GetDoubleValue(), got[i].GetDoubleValue())
			}
		}

	}
}

func compareLabelVals(t *testing.T, want []*v1.LabelValue, got []*v1.LabelValue) {
	if len(want) != len(got) {
		t.Errorf("Metric to Label Values failed, expected length %v, got %v", len(want), len(got))
	}

	for i := range want {
		if want[i].GetValue() != got[i].GetValue() {
			t.Errorf("Metric to Label Values failed, expected val %v, got %v", want[i].GetValue(), got[i].GetValue())
		}

		if want[i].GetHasValue() != got[i].GetHasValue() {
			t.Errorf("Metric to Label Values failed, expected has value %v, got %v", want[i].GetHasValue(), got[i].GetHasValue())
		}
	}
}

func compareMetricTimeseries(t *testing.T, want []*v1.TimeSeries, got []*v1.TimeSeries) {
	for i := range want {
		if want[i].GetStartTimestamp().String() != got[i].GetStartTimestamp().String() {
			t.Errorf("Metric to timeseries failed, expected val %v, got %v", want[i].GetStartTimestamp().String(), got[i].GetStartTimestamp().String())
		}

		compareLabelVals(t, want[i].LabelValues, got[i].LabelValues)
		comparePoints(t, want[i].Points, got[i].Points, true)
	}
}

func compareLabelKeys(t *testing.T, want []*v1.LabelKey, got []*v1.LabelKey) {
	for i := range labelKeys {
		if labelKeys[i].GetKey() != got[i].GetKey() {
			t.Errorf("Metric to Label Key failed, expected key %v, got %v", want[i].GetKey(), got[i].GetKey())
		}

		if labelKeys[i].GetDescription() != got[i].GetDescription() {
			t.Errorf("Metric to Label Key failed, expected desc %v, got %v", want[i].GetDescription(), got[i].GetDescription())
		}
	}
}

func compareMetricDesc(t *testing.T, want *v1.MetricDescriptor, got *v1.MetricDescriptor) {
	if want.Name != got.Name {
		t.Errorf("Metric to Descriptor failed, expected name %v, got %v", want.Name, got.Name)
	}

	if want.Description != got.Description {
		t.Errorf("Metric to Descriptor failed, expected description %v, got %v", want.Description, got.Description)
	}

	if want.Unit != got.Unit {
		t.Errorf("Metric to Descriptor failed, expected unit %v, got %v", want.Unit, got.Unit)
	}

	if want.Type != got.Type {
		t.Errorf("Metric to Descriptor failed, expected type %v, got %v", want.Type, got.Type)
	}

	compareLabelKeys(t, want.LabelKeys, got.LabelKeys)
}
