// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package pcfmetrics

import (
	"fmt"
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

type counter interface {
	Count() int64
}

type gauge interface {
	Value() int64
}

type gaugeFloat64 interface {
	Value() float64
}

type meter interface {
	Count() int64
	Rate1() float64
	Rate5() float64
	Rate15() float64
	RateMean() float64
}

type histogram interface {
	Count() int64
	Sum() int64
	Max() int64
	Min() int64
	Mean() float64
	StdDev() float64
	Variance() float64
	Percentiles([]float64) []float64
}

type timer interface {
	Count() int64
	Rate1() float64
	Rate5() float64
	Rate15() float64
	RateMean() float64
	Sum() int64
	Max() int64
	Min() int64
	Mean() float64
	StdDev() float64
	Variance() float64
	Percentiles([]float64) []float64
}

func convertGauge(gauge gauge, name string) *dataPoint {
	return convertGenericGauge(float64(gauge.Value()), name)
}

func convertGaugeFloat64(gauge gaugeFloat64, name string) *dataPoint {
	return convertGenericGauge(gauge.Value(), name)
}

func convertMeter(meter meter, name string) []*dataPoint {
	return []*dataPoint{
		convertCounter(meter, joinNameParts(name, "count")),
		convertGenericGauge(meter.Rate1(), joinNameParts(name, "rate.1-minute")),
		convertGenericGauge(meter.Rate5(), joinNameParts(name, "rate.5-minute")),
		convertGenericGauge(meter.Rate15(), joinNameParts(name, "rate.15-minute")),
		convertGenericGauge(meter.RateMean(), joinNameParts(name, "rate.mean")),
	}
}

func convertHistogram(histogram histogram, name string) []*dataPoint {
	return convertHistogramWithTimeUnit(histogram, name, time.Duration(0))
}

func convertHistogramWithTimeUnit(histogram histogram, name string, timeUnit time.Duration) []*dataPoint {
	points := []*dataPoint{
		convertCounter(histogram, joinNameParts(name, "count")),
		convertGenericGaugeWithUnit(histogram.Mean(), joinNameParts(name, "mean"), timeUnit),
		convertGenericGaugeWithUnit(histogram.StdDev(), joinNameParts(name, "stddev"), timeUnit),
		convertGenericGaugeWithUnit(float64(histogram.Sum()), joinNameParts(name, "sum"), timeUnit),
		convertGenericGaugeWithUnit(histogram.Variance(), joinNameParts(name, "variance"), timeUnit),
		convertGenericGaugeWithUnit(float64(histogram.Max()), joinNameParts(name, "max"), timeUnit),
		convertGenericGaugeWithUnit(float64(histogram.Min()), joinNameParts(name, "min"), timeUnit),
	}

	points = append(points, generatePercentileDataPoints(histogram, name, timeUnit)...)

	return points
}

func generatePercentileDataPoints(histogram histogram, name string, timeUnit time.Duration) []*dataPoint {
	var points []*dataPoint
	percentileIds := []float64{75, 95, 98, 99, 99.9}
	for i, value := range histogram.Percentiles(percentileIds) {
		dataPoint := convertGenericGaugeWithUnit(
			float64(value),
			getPercentileName(name, percentileIds[i]),
			timeUnit,
		)
		points = append(points, dataPoint)
	}

	return points
}

func getPercentileName(name string, percentileId float64) string {
	percentileIdString := strconv.FormatFloat(percentileId, 'f', -1, 64)
	percentileWithoutPeriods := strings.Replace(percentileIdString, ".", "", -1)
	return joinNameParts(name, fmt.Sprintf("%sthPercentile", percentileWithoutPeriods))
}

func convertTimer(timer timer, name string, timeUnit time.Duration) []*dataPoint {
	points := []*dataPoint{
		convertCounter(timer, joinNameParts(name, "count")),
	}

	meterDataPoints := convertMeter(timer, name)
	meterDataPointsWithoutCounter := meterDataPoints[1:]
	points = append(points, meterDataPointsWithoutCounter...)

	if timeUnit == time.Duration(0) {
		timeUnit = time.Millisecond
	}

	histogramDataPoints := convertHistogramWithTimeUnit(timer, joinNameParts(name, "duration"), timeUnit)
	histogramDataPointsWithoutCounter := histogramDataPoints[1:]
	points = append(points, histogramDataPointsWithoutCounter...)

	return points
}

func convertCounter(counter counter, name string) *dataPoint {
	return &dataPoint{
		Name:      name,
		Value:     float64(counter.Count()),
		Type:      "counter",
	}
}

func convertGenericGauge(value float64, name string) *dataPoint {
	return &dataPoint{
		Name:      name,
		Value:     value,
		Type:      "gauge",
	}
}

func convertGenericGaugeWithUnit(value float64, name string, timeUnit time.Duration) *dataPoint {
	if timeUnit == time.Duration(0) {
		return convertGenericGauge(value, name)
	}

	return &dataPoint{
		Name:      name,
		Value:     value / float64(timeUnit),
		Type:      "gauge",
		Unit:      getTimeUnitName(timeUnit),
	}
}

func getTimeUnitName(timeUnit time.Duration) string {
	switch {
	case timeUnit == time.Second:
		return "seconds"
	case timeUnit == time.Millisecond:
		return "milliseconds"
	case timeUnit == time.Microsecond:
		return "microseconds"
	case timeUnit == time.Nanosecond:
		return "nanoseconds"
	default:
		return "milliseconds"
	}
}

func joinNameParts(names ...string) string {
	return strings.Join(names, ".")
}
