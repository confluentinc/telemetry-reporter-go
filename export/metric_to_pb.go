package export

import (
	"fmt"

	a1 "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	v1 "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	r1 "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/resource"
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
		Resource:         resourceToProto(m.Resource),
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
		Type:        metricdataTypetoProtoType(m.Descriptor.Type),
		LabelKeys:   metricToLabelKeys(m),
	}
}

/*
We need this function since OpenCensus Metric Enum types are offset by one in
the protobuf format
see https://github.com/census-instrumentation/opencensus-go/blob/master/metric/metricdata/point.go#L185-L193
and https://github.com/census-instrumentation/opencensus-proto/blob/master/gen-go/metrics/v1/metrics.pb.go#L61-L89
*/
func metricdataTypetoProtoType(metricType metricdata.Type) v1.MetricDescriptor_Type {
	return v1.MetricDescriptor_Type(metricType + 1)
}

func resourceToProto(r *resource.Resource) *r1.Resource {
	if r != nil {
		return &r1.Resource{
			Type:   r.Type,
			Labels: r.Labels,
		}
	}

	return nil

}

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
		case *metricdata.Summary:
			toAppend.Value = &v1.Point_SummaryValue{
				SummaryValue: pointToSummaryValue(v),
			}
		case *metricdata.Distribution:
			toAppend.Value = &v1.Point_DistributionValue{
				DistributionValue: pointToDistributionValue(v),
			}
		default:
			panic("unsupported value type")
		}

		points = append(points, toAppend)
	}

	return points
}

func pointToSummaryValue(value *metricdata.Summary) *v1.SummaryValue {
	return &v1.SummaryValue{
		Count:    &wrappers.Int64Value{Value: value.Count},
		Sum:      &wrappers.DoubleValue{Value: value.Sum},
		Snapshot: summaryValToSnapshot(value),
	}
}

func summaryValToSnapshot(summary *metricdata.Summary) *v1.SummaryValue_Snapshot {
	snapshot := &v1.SummaryValue_Snapshot{}

	if summary.HasCountAndSum {
		snapshot.Count = &wrappers.Int64Value{Value: summary.Snapshot.Count}
		snapshot.Sum = &wrappers.DoubleValue{Value: summary.Snapshot.Sum}
	}

	percentileValues := []*v1.SummaryValue_Snapshot_ValueAtPercentile{}
	for percentile, val := range summary.Snapshot.Percentiles {
		toAppend := &v1.SummaryValue_Snapshot_ValueAtPercentile{
			Percentile: percentile,
			Value:      val,
		}
		percentileValues = append(percentileValues, toAppend)
	}

	snapshot.PercentileValues = percentileValues
	return snapshot
}

func pointToDistributionValue(value *metricdata.Distribution) *v1.DistributionValue {
	return &v1.DistributionValue{
		Count:                 value.Count,
		Sum:                   value.Sum,
		SumOfSquaredDeviation: value.SumOfSquaredDeviation,
		Buckets:               distributionToBuckets(value),
		BucketOptions:         distributionToBucketOptions(value),
	}
}

func distributionToBuckets(value *metricdata.Distribution) []*v1.DistributionValue_Bucket {
	buckets := []*v1.DistributionValue_Bucket{}

	for _, bucket := range value.Buckets {
		buckets = append(buckets, bucketToProto(bucket))
	}

	return buckets
}

func bucketToProto(bucket metricdata.Bucket) *v1.DistributionValue_Bucket {
	res := &v1.DistributionValue_Bucket{
		Count: bucket.Count,
	}

	if bucket.Exemplar != nil {
		res.Exemplar = bucketToExemplar(bucket)
	}

	return res
}

func bucketToExemplar(bucket metricdata.Bucket) *v1.DistributionValue_Exemplar {
	timestamp, _ := ptypes.TimestampProto(bucket.Exemplar.Timestamp)

	return &v1.DistributionValue_Exemplar{
		Value:       bucket.Exemplar.Value,
		Timestamp:   timestamp,
		Attachments: bucketToAttachments(bucket),
	}
}

func bucketToAttachments(bucket metricdata.Bucket) map[string]string {
	res := map[string]string{}

	for k, v := range bucket.Exemplar.Attachments {
		res[k] = fmt.Sprintf("%v", v)
	}

	return res
}

func distributionToBucketOptions(value *metricdata.Distribution) *v1.DistributionValue_BucketOptions {
	return &v1.DistributionValue_BucketOptions{
		Type: &v1.DistributionValue_BucketOptions_Explicit_{
			Explicit: &v1.DistributionValue_BucketOptions_Explicit{
				Bounds: value.BucketOptions.Bounds,
			},
		},
	}
}
