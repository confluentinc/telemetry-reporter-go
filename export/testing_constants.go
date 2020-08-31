package export

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"go.opencensus.io/metric/metricdata"
)

var (
	dummyIncludeFilter   = `.*`
	dummyReportingPeriod = 1000

	config = Config{
		IncludeFilter:               dummyIncludeFilter,
		reportingPeriodMilliseconds: dummyReportingPeriod,
	}
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

	metric = &metricdata.Metric{
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

	metrics = []*metricdata.Metric{metric}
)
