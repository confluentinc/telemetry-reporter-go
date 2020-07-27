package export

import (
	a1 "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	v1 "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/golang/protobuf/ptypes"
	"go.opencensus.io/metric/metricdata"
)

func metricsToServiceRequest(ms []*metricdata.Metric) *a1.ExportMetricsServiceRequest {
	metrics := []*v1.Metric{}

	for _, m := range ms {
		toAppend := metricToProto(m)
		metrics = append(metrics, toAppend)
	}

	return &a1.ExportMetricsServiceRequest{
		Metrics: metrics,
	}
}

func metricToProto(m *metricdata.Metric) *v1.Metric {
	return &v1.Metric{
		MetricDescriptor: metricToDescriptor(m),
		Timeseries:       metricToTimeSeries(m),
		// Resource:         metricToResource(m),
	}
}

func metricToLabelKeys(m *metricdata.Metric) []*v1.LabelKey {
	labelKeys := []*v1.LabelKey{}

	for _, lk := range m.Descriptor.LabelKeys {
		toAppend := &v1.LabelKey{
			Key:         lk.Key,
			Description: lk.Description,
		}

		labelKeys = append(labelKeys, toAppend)
	}
	return labelKeys
}

func metricToDescriptor(m *metricdata.Metric) *v1.MetricDescriptor {
	return &v1.MetricDescriptor{
		Name:        m.Descriptor.Name,
		Description: m.Descriptor.Description,
		Unit:        string(m.Descriptor.Unit),
		Type:        v1.MetricDescriptor_Type(m.Descriptor.Type + 1),
		LabelKeys:   metricToLabelKeys(m),
	}
}

// func metricToResource(m *metricdata.Metric) *r1.Resource {
// 	return &r1.Resource{
// 		Type:   m.Resource.Type,
// 		Labels: m.Resource.Labels,
// 	}
// }

func metricToTimeSeries(m *metricdata.Metric) []*v1.TimeSeries {
	timeSeries := []*v1.TimeSeries{}

	for _, ts := range m.TimeSeries {
		timestamp, _ := ptypes.TimestampProto(ts.StartTime)

		toAppend := &v1.TimeSeries{
			StartTimestamp: timestamp,
			LabelValues:    metricToLabelValues(ts),
			Points:         metricToPoints(ts),
		}

		timeSeries = append(timeSeries, toAppend)
	}

	return timeSeries
}

func metricToLabelValues(t *metricdata.TimeSeries) []*v1.LabelValue {
	labelValues := []*v1.LabelValue{}

	for _, lv := range t.LabelValues {
		toAppend := &v1.LabelValue{
			Value:    lv.Value,
			HasValue: lv.Present,
		}

		labelValues = append(labelValues, toAppend)
	}

	return labelValues
}

func metricToPoints(t *metricdata.TimeSeries) []*v1.Point {
	points := []*v1.Point{}

	for _, p := range t.Points {
		timestamp, _ := ptypes.TimestampProto(p.Time)
		toAppend := &v1.Point{
			Timestamp: timestamp,
		}

		switch v := p.Value.(type) {
		case int64:
			toAppend.Value = &v1.Point_Int64Value{
				Int64Value: v,
			}
		case float64:
			toAppend.Value = &v1.Point_DoubleValue{
				DoubleValue: v,
			}
		// TODO Distribution and Summary Value
		default:
			panic("unsupported value type")
		}

		points = append(points, toAppend)
	}

	return points
}
