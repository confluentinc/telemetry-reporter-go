package export

import (
	"reflect"
	"testing"

	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/resource"

	a1 "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	v1 "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	r1 "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"github.com/golang/protobuf/ptypes/wrappers"
)

var (
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

	summaryPoints = []*v1.Point{
		&v1.Point{
			Timestamp: timestamp,
			Value: &v1.Point_SummaryValue{
				SummaryValue: &v1.SummaryValue{
					Count: &wrappers.Int64Value{Value: intVal},
					Sum:   &wrappers.DoubleValue{Value: doubleVal},
					Snapshot: &v1.SummaryValue_Snapshot{
						Count:            &wrappers.Int64Value{Value: 0},
						Sum:              &wrappers.DoubleValue{Value: 0},
						PercentileValues: []*v1.SummaryValue_Snapshot_ValueAtPercentile{},
					},
				},
			},
		},
	}

	labelValsProto = []*v1.LabelValue{
		&v1.LabelValue{
			Value:    dummyLabelVal,
			HasValue: true,
		},
	}

	timeseriesProto = []*v1.TimeSeries{
		&v1.TimeSeries{
			StartTimestamp: timestamp,
			LabelValues:    labelValsProto,
			Points:         intPoints,
		},
	}

	labelKeysProto = []*v1.LabelKey{
		&v1.LabelKey{
			Key:         dummyLabelKey,
			Description: dummyKeyDesc,
		},
	}

	descriptorProto = &v1.MetricDescriptor{
		Name:        dummyName,
		Description: dummyDesc,
		Unit:        dummyUnit,
		Type:        v1.MetricDescriptor_Type(dummyType + 1),
		LabelKeys:   labelKeysProto,
	}

	dummyMetricProto = &v1.Metric{
		MetricDescriptor: descriptorProto,
		Timeseries:       timeseriesProto,
	}

	dummyServiceRequestProto = &a1.ExportMetricsServiceRequest{
		Metrics: []*v1.Metric{
			dummyMetricProto,
		},
	}
)

func TestMetricToPointInt64(t *testing.T) {
	timeseries := metricdata.TimeSeries{
		Points: []metricdata.Point{metricdata.NewInt64Point(timeNow, intVal)},
	}
	got, err := metricToPoints(&timeseries)
	if err != nil {
		t.Errorf("Error converting metric to timeseries proto: %v", err)
	}
	comparePoints(t, intPoints, got)
}

func TestMetricToPointDouble64(t *testing.T) {
	timeseries := metricdata.TimeSeries{
		Points: []metricdata.Point{metricdata.NewFloat64Point(timeNow, doubleVal)},
	}
	got, err := metricToPoints(&timeseries)
	if err != nil {
		t.Errorf("Error converting metric to timeseries proto: %v", err)
	}
	comparePoints(t, doublePoints, got)
}

func TestMetricToPointSummary(t *testing.T) {
	snapshot := metricdata.Snapshot{
		Count:       0,
		Sum:         0,
		Percentiles: map[float64]float64{},
	}

	sumVal := &metricdata.Summary{
		Count:          intVal,
		Sum:            doubleVal,
		HasCountAndSum: true,
		Snapshot:       snapshot,
	}

	timeseries := metricdata.TimeSeries{
		Points: []metricdata.Point{metricdata.NewSummaryPoint(timeNow, sumVal)},
	}

	got, err := metricToPoints(&timeseries)
	if err != nil {
		t.Errorf("Error converting metric to timeseries proto: %v", err)
	}
	comparePoints(t, summaryPoints, got)
}

func TestMetricToLabelValues(t *testing.T) {
	got := metricToLabelValues(metric.TimeSeries[0])
	compareLabelVals(t, labelValsProto, got)
}

func TestMetricToTimeSeries(t *testing.T) {
	got, err := metricToTimeSeries(metric)
	if err != nil {
		t.Errorf("Error converting metric to timeseries proto: %v", err)
	}
	compareMetricTimeseries(t, timeseriesProto, got)
}

func TestResourceToProto(t *testing.T) {
	r := &resource.Resource{
		Type: "test",
		Labels: map[string]string{
			"key": "val",
		},
	}

	got := resourceToProto(r)
	want := &r1.Resource{
		Type: "test",
		Labels: map[string]string{
			"key": "val",
		},
	}

	compareResources(t, want, got)
}

func TestMetricToDescriptor(t *testing.T) {
	got := metricToDescriptor(metric)
	compareMetricDesc(t, descriptorProto, got)
}

func TestMetricToLabelKeys(t *testing.T) {
	got := metricToLabelKeys(metric)
	compareLabelKeys(t, labelKeysProto, got)
}

func TestMetricToProto(t *testing.T) {
	got, err := metricToProto(metric)
	if err != nil {
		t.Errorf("Error converting metric to proto")
	}
	compareMetricTimeseries(t, dummyMetricProto.Timeseries, got.Timeseries)
	compareMetricDesc(t, dummyMetricProto.GetMetricDescriptor(), got.GetMetricDescriptor())
}

func TestMetricToServiceRequest(t *testing.T) {
	got, err := metricsToServiceRequest(metrics)
	if err != nil {
		t.Errorf("Error converting metrics to service proto ")
	}
	for i := range dummyServiceRequestProto.Metrics {
		compareMetricTimeseries(t, dummyServiceRequestProto.Metrics[i].Timeseries, got.Metrics[i].Timeseries)
		compareMetricDesc(t, dummyServiceRequestProto.Metrics[i].GetMetricDescriptor(), got.Metrics[i].GetMetricDescriptor())
	}
}

func comparePoints(t *testing.T, want []*v1.Point, got []*v1.Point) {
	if len(want) != len(got) {
		t.Errorf("Metric to Points int failed, expected length %v, got %v", len(want), len(got))
	}

	for i := range want {
		if want[i].Timestamp.String() != got[i].Timestamp.String() {
			t.Errorf("Metric to Points int failed, expected time %v, got %v", want[i].Timestamp, got[i].Timestamp)
		}

		switch val := want[i].Value.(type) {
		case *v1.Point_Int64Value:
			if val.Int64Value != got[i].GetInt64Value() {
				t.Errorf("Metric to Points int failed, expected val %v, got %v", val.Int64Value, got[i].GetInt64Value())
			}
		case *v1.Point_DoubleValue:
			if val.DoubleValue != got[i].GetDoubleValue() {
				t.Errorf("Metric to Points double failed, expected val %v, got %v", val.DoubleValue, got[i].GetDoubleValue())
			}
		case *v1.Point_SummaryValue:
			if !reflect.DeepEqual(*want[i].GetSummaryValue(), *got[i].GetSummaryValue()) {
				t.Errorf("Metric to Points summary failed, expected val %v, got %v", val.SummaryValue, got[i].GetSummaryValue())
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
		comparePoints(t, want[i].Points, got[i].Points)
	}
}

func compareLabelKeys(t *testing.T, want []*v1.LabelKey, got []*v1.LabelKey) {
	for i := range want {
		if want[i].GetKey() != got[i].GetKey() {
			t.Errorf("Metric to Label Key failed, expected key %v, got %v", want[i].GetKey(), got[i].GetKey())
		}

		if want[i].GetDescription() != got[i].GetDescription() {
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

func compareResources(t *testing.T, want *r1.Resource, got *r1.Resource) {
	if want.Type != got.Type {
		t.Errorf("Resource to Proto failed, expected type %v, got %v", want.Type, got.Type)
	}

	if !reflect.DeepEqual(want.Labels, got.Labels) {
		t.Errorf("Resource to Proto failed, expected labels %v, labels %v", want.Labels, got.Labels)
	}
}
