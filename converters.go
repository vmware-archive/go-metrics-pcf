package pcf

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"strconv"
	"strings"
	"time"
)

//go:generate Counterfeiter github.com/rcrowley/go-metrics.Counter
//go:generate Counterfeiter github.com/rcrowley/go-metrics.Gauge
//go:generate Counterfeiter github.com/rcrowley/go-metrics.GaugeFloat64
//go:generate Counterfeiter github.com/rcrowley/go-metrics.Meter
//go:generate Counterfeiter github.com/rcrowley/go-metrics.Histogram
//go:generate Counterfeiter github.com/rcrowley/go-metrics.Timer
//go:generate Counterfeiter github.com/rcrowley/go-metrics.EWMA

func convertCounter(counter metrics.Counter, name string, currentTime int64) *dataPoint {
	return &dataPoint{
		Name:      name,
		Value:     float64(counter.Count()),
		Timestamp: currentTime,
		Type:      "counter",
	}
}

func convertGauge(gauge metrics.Gauge, name string, currentTime int64) *dataPoint {
	return &dataPoint{
		Name:      name,
		Value:     float64(gauge.Value()),
		Timestamp: currentTime,
		Type:      "gauge",
	}
}

func convertGaugeFloat64(gauge metrics.GaugeFloat64, name string, currentTime int64) *dataPoint {
	return &dataPoint{
		Name:      name,
		Value:     float64(gauge.Value()),
		Timestamp: currentTime,
		Type:      "gauge",
	}
}

func convertMeter(meter metrics.Meter, name string, currentTime int64) []*dataPoint {
	return []*dataPoint{
		{
			Name:      namer(name, "count"),
			Value:     float64(meter.Count()),
			Timestamp: currentTime,
			Type:      "counter",
		},
		{
			Name:      namer(name, "rate.1-minute"),
			Value:     float64(meter.Rate1()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "rate.5-minute"),
			Value:     float64(meter.Rate5()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "rate.15-minute"),
			Value:     float64(meter.Rate15()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "rate.mean"),
			Value:     float64(meter.RateMean()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
	}
}

func convertHistogram(histogram metrics.Histogram, name string, currentTime int64) []*dataPoint {
	points := []*dataPoint{
		{
			Name:      namer(name, "count"),
			Value:     float64(histogram.Count()),
			Timestamp: currentTime,
			Type:      "counter",
		},
		{
			Name:      namer(name, "mean"),
			Value:     float64(histogram.Mean()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "stddev"),
			Value:     float64(histogram.StdDev()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "sum"),
			Value:     float64(histogram.Sum()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "variance"),
			Value:     float64(histogram.Variance()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "max"),
			Value:     float64(histogram.Max()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "min"),
			Value:     float64(histogram.Min()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
	}

	percentiles := []float64{75, 95, 98, 99, 99.9}
	for i, v := range histogram.Percentiles(percentiles) {
		percentileName := strings.Replace(strconv.FormatFloat(percentiles[i], 'f', -1, 64), ".", "", -1)
		points = append(points, &dataPoint{
			Name:      namer(name, fmt.Sprintf("%sthPercentile", percentileName)),
			Value:     float64(v),
			Timestamp: currentTime,
			Type:      "gauge",
		})
	}

	return points
}

func convertTimer(timer metrics.Timer, name string, currentTime int64, timeUnit time.Duration) []*dataPoint {

	unit := ""

	switch {
	case timeUnit == time.Second:
		unit = "seconds"
	case timeUnit == time.Millisecond:
		unit = "milliseconds"
	case timeUnit == time.Microsecond:
		unit = "microseconds"
	case timeUnit == time.Nanosecond:
		unit = "nanoseconds"
	}

	points := []*dataPoint{
		{
			Name:      namer(name, "count"),
			Value:     float64(timer.Count()),
			Timestamp: currentTime,
			Type:      "counter",
		},
		{
			Name:      namer(name, "rate.1-minute"),
			Value:     float64(timer.Rate1()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "rate.5-minute"),
			Value:     float64(timer.Rate5()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "rate.15-minute"),
			Value:     float64(timer.Rate15()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "rate.mean"),
			Value:     float64(timer.RateMean()),
			Timestamp: currentTime,
			Type:      "gauge",
		},
		{
			Name:      namer(name, "duration.mean"),
			Value:     timer.Mean() / float64(timeUnit),
			Timestamp: currentTime,
			Type:      "gauge",
			Unit: unit,
		},
		{
			Name:      namer(name, "duration.stddev"),
			Value:     timer.StdDev() / float64(timeUnit),
			Timestamp: currentTime,
			Type:      "gauge",
			Unit: unit,
		},
		{
			Name:      namer(name, "duration.sum"),
			Value:     float64(timer.Sum() / int64(timeUnit)),
			Timestamp: currentTime,
			Type:      "gauge",
			Unit: unit,
		},
		{
			Name:      namer(name, "duration.variance"),
			Value:     timer.Variance() / float64(timeUnit),
			Timestamp: currentTime,
			Type:      "gauge",
			Unit: unit,
		},
		{
			Name:      namer(name, "duration.max"),
			Value:     float64(timer.Max() / int64(timeUnit)),
			Timestamp: currentTime,
			Type:      "gauge",
			Unit: unit,
		},
		{
			Name:      namer(name, "duration.min"),
			Value:     float64(timer.Min() / int64(timeUnit)),
			Timestamp: currentTime,
			Type:      "gauge",
			Unit: unit,
		},
	}

	percentiles := []float64{75, 95, 98, 99, 99.9}
	for i, v := range timer.Percentiles(percentiles) {
		percentileName := strings.Replace(strconv.FormatFloat(percentiles[i], 'f', -1, 64), ".", "", -1)
		points = append(points, &dataPoint{
			Name:      namer(name, "duration", fmt.Sprintf("%sthPercentile", percentileName)),
			Value:     v / float64(timeUnit),
			Timestamp: currentTime,
			Type:      "gauge",
			Unit: unit,
		})
	}

	return points
}

func namer(names ...string) string {
	return strings.Join(names, ".")
}
