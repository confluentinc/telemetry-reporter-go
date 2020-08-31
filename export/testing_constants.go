package export

import "go.opencensus.io/metric/metricdata"

var (
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
