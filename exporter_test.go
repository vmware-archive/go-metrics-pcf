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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metricFakes "github.com/pivotal-cf/go-metrics-pcf/go-metrics-pcffakes"
	"github.com/rcrowley/go-metrics"
)

var _ = Describe("`go-metrics` exporter for PCF Metrics", func() {
	type testContext struct {
		registry          metrics.Registry
		transportMessages chan []*dataPoint
		fakeTimeHelper    *fakeTimeHelper
		exporter          *exporter
	}

	var setup = func() *testContext {
		tc := &testContext{
			registry:          metrics.NewRegistry(),
			transportMessages: make(chan []*dataPoint, 100),
			fakeTimeHelper:    newFakeTimeHelper(),
		}

		tc.exporter = newExporter(newFakeTransporter(tc.transportMessages), tc.fakeTimeHelper, time.Millisecond)
		return tc
	}

	It("exports counter metrics", func() {
		tc := setup()
		tc.fakeTimeHelper.returnValue = 123

		fakeCounter := new(metricFakes.FakeCounter)
		fakeCounter.SnapshotReturns(fakeCounter)
		fakeCounter.CountReturns(6)

		tc.registry.Register("test-counter", fakeCounter)
		tc.exporter.exportMetrics(tc.registry)

		Expect(fakeCounter.SnapshotCallCount()).To(Equal(1))
		Expect(tc.transportMessages).To(Receive(ConsistOf(
			&dataPoint{
				Name:      "test-counter",
				Type:      "counter",
				Value:     6,
				Timestamp: 123,
				Unit:      "",
			},
		)))
	})

	It("exports gauge metrics", func() {
		tc := setup()
		tc.fakeTimeHelper.returnValue = 123

		fakeGauge := new(metricFakes.FakeGauge)
		fakeGauge.SnapshotReturns(fakeGauge)
		fakeGauge.ValueReturns(17)

		tc.registry.Register("test-gauge", fakeGauge)
		tc.exporter.exportMetrics(tc.registry)

		Expect(fakeGauge.SnapshotCallCount()).To(Equal(1))
		Expect(tc.transportMessages).To(Receive(ConsistOf(
			&dataPoint{
				Name:      "test-gauge",
				Type:      "gauge",
				Value:     17,
				Timestamp: 123,
				Unit:      "",
			},
		)))
	})

	It("exports gauge float64 metrics", func() {
		tc := setup()
		tc.fakeTimeHelper.returnValue = 123

		fakeGaugeFloat64 := new(metricFakes.FakeGaugeFloat64)
		fakeGaugeFloat64.SnapshotReturns(fakeGaugeFloat64)
		fakeGaugeFloat64.ValueReturns(32.2)

		tc.registry.Register("test-gauge-float-64", fakeGaugeFloat64)
		tc.exporter.exportMetrics(tc.registry)

		Expect(fakeGaugeFloat64.SnapshotCallCount()).To(Equal(1))
		Expect(tc.transportMessages).To(Receive(ConsistOf(
			&dataPoint{
				Name:      "test-gauge-float-64",
				Type:      "gauge",
				Value:     32.2,
				Timestamp: 123,
				Unit:      "",
			},
		)))
	})

	It("exports meter metrics", func() {
		tc := setup()
		tc.fakeTimeHelper.returnValue = 123

		fakeMeter := new(metricFakes.FakeMeter)
		fakeMeter.SnapshotReturns(fakeMeter)
		fakeMeter.CountReturns(1)
		fakeMeter.Rate1Returns(2)
		fakeMeter.Rate5Returns(3)
		fakeMeter.Rate15Returns(4)
		fakeMeter.RateMeanReturns(5)

		tc.registry.Register("test-fakeMeter", fakeMeter)
		tc.exporter.exportMetrics(tc.registry)

		Expect(fakeMeter.SnapshotCallCount()).To(Equal(1))

		points := []*dataPoint{}
		Expect(tc.transportMessages).To(Receive(&points))
		Expect(points).To(HaveLen(5))

		Expect(points).To(ContainElement(
			&dataPoint{
				Name:      "test-fakeMeter.count",
				Type:      "counter",
				Value:     1,
				Timestamp: 123,
				Unit:      "",
			},
		))

		Expect(points).To(ContainElement(
			&dataPoint{
				Name:      "test-fakeMeter.rate.1-minute",
				Type:      "gauge",
				Value:     2,
				Timestamp: 123,
				Unit:      "",
			},
		))

		Expect(points).To(ContainElement(
			&dataPoint{
				Name:      "test-fakeMeter.rate.5-minute",
				Type:      "gauge",
				Value:     3,
				Timestamp: 123,
				Unit:      "",
			},
		))

		Expect(points).To(ContainElement(
			&dataPoint{
				Name:      "test-fakeMeter.rate.15-minute",
				Type:      "gauge",
				Value:     4,
				Timestamp: 123,
				Unit:      "",
			},
		))

		Expect(points).To(ContainElement(
			&dataPoint{
				Name:      "test-fakeMeter.rate.mean",
				Type:      "gauge",
				Value:     5,
				Timestamp: 123,
				Unit:      "",
			},
		))
	})

	It("exports histogram metrics", func() {
		tc := setup()
		tc.fakeTimeHelper.returnValue = 123

		fakeHistogram := new(metricFakes.FakeHistogram)
		fakeHistogram.SnapshotReturns(fakeHistogram)
		fakeHistogram.CountReturns(1)
		fakeHistogram.MeanReturns(2)
		fakeHistogram.StdDevReturns(3)
		fakeHistogram.SumReturns(4)
		fakeHistogram.VarianceReturns(5)
		fakeHistogram.MaxReturns(6)
		fakeHistogram.MinReturns(7)
		fakeHistogram.PercentilesReturns([]float64{8, 9, 10, 11, 12})

		tc.registry.Register("test-histogram", fakeHistogram)
		tc.exporter.exportMetrics(tc.registry)

		Expect(fakeHistogram.SnapshotCallCount()).To(Equal(1))

		Expect(tc.transportMessages).To(Receive(ConsistOf(
			&dataPoint{
				Name:      "test-histogram.count",
				Type:      "counter",
				Value:     1,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.mean",
				Type:      "gauge",
				Value:     2,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.stddev",
				Type:      "gauge",
				Value:     3,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.sum",
				Type:      "gauge",
				Value:     4,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.variance",
				Type:      "gauge",
				Value:     5,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.max",
				Type:      "gauge",
				Value:     6,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.min",
				Type:      "gauge",
				Value:     7,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.75thPercentile",
				Type:      "gauge",
				Value:     8,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.95thPercentile",
				Type:      "gauge",
				Value:     9,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.98thPercentile",
				Type:      "gauge",
				Value:     10,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.99thPercentile",
				Type:      "gauge",
				Value:     11,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-histogram.999thPercentile",
				Type:      "gauge",
				Value:     12,
				Timestamp: 123,
				Unit:      "",
			},
		)))

		Expect(fakeHistogram.PercentilesArgsForCall(0)).To(Equal([]float64{75, 95, 98, 99, 99.9}))
	})

	It("exports timer metrics", func() {
		tc := setup()
		tc.fakeTimeHelper.returnValue = 123

		fakeTimer := new(metricFakes.FakeTimer)
		fakeTimer.SnapshotReturns(fakeTimer)
		fakeTimer.CountReturns(1)
		fakeTimer.Rate1Returns(2)
		fakeTimer.Rate5Returns(3)
		fakeTimer.Rate15Returns(4)
		fakeTimer.RateMeanReturns(16)
		fakeTimer.MeanReturns(5 * float64(time.Millisecond))
		fakeTimer.StdDevReturns(6 * float64(time.Millisecond))
		fakeTimer.SumReturns(7 * int64(time.Millisecond))
		fakeTimer.VarianceReturns(8 * float64(time.Millisecond))
		fakeTimer.MaxReturns(9 * int64(time.Millisecond))
		fakeTimer.MinReturns(10 * int64(time.Millisecond))
		fakeTimer.PercentilesReturns([]float64{
			11 * float64(time.Millisecond),
			12 * float64(time.Millisecond),
			13 * float64(time.Millisecond),
			14 * float64(time.Millisecond),
			15 * float64(time.Millisecond),
		})

		tc.registry.Register("test-timer", fakeTimer)
		tc.exporter.exportMetrics(tc.registry)

		Expect(fakeTimer.SnapshotCallCount()).To(Equal(1))

		Expect(tc.transportMessages).To(Receive(ConsistOf(
			&dataPoint{
				Name:      "test-timer.count",
				Type:      "counter",
				Value:     1,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-timer.rate.1-minute",
				Type:      "gauge",
				Value:     2,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-timer.rate.5-minute",
				Type:      "gauge",
				Value:     3,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-timer.rate.15-minute",
				Type:      "gauge",
				Value:     4,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-timer.rate.mean",
				Type:      "gauge",
				Value:     16,
				Timestamp: 123,
				Unit:      "",
			},
			&dataPoint{
				Name:      "test-timer.duration.mean",
				Type:      "gauge",
				Value:     5,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.stddev",
				Type:      "gauge",
				Value:     6,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.sum",
				Type:      "gauge",
				Value:     7,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.variance",
				Type:      "gauge",
				Value:     8,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.max",
				Type:      "gauge",
				Value:     9,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.min",
				Type:      "gauge",
				Value:     10,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.75thPercentile",
				Type:      "gauge",
				Value:     11,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.95thPercentile",
				Type:      "gauge",
				Value:     12,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.98thPercentile",
				Type:      "gauge",
				Value:     13,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.99thPercentile",
				Type:      "gauge",
				Value:     14,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
			&dataPoint{
				Name:      "test-timer.duration.999thPercentile",
				Type:      "gauge",
				Value:     15,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
		)))

		Expect(fakeTimer.PercentilesArgsForCall(0)).To(Equal([]float64{75, 95, 98, 99, 99.9}))
	})

	It("uses milliseconds as the default time unit", func() {
		tc := &testContext{
			registry:          metrics.NewRegistry(),
			transportMessages: make(chan []*dataPoint, 100),
			fakeTimeHelper:    newFakeTimeHelper(),
		}

		options := &Options{}
		tc.exporter = newExporter(newFakeTransporter(tc.transportMessages), tc.fakeTimeHelper, options.TimeUnit)
		tc.fakeTimeHelper.returnValue = 123

		fakeTimer := new(metricFakes.FakeTimer)
		fakeTimer.SnapshotReturns(fakeTimer)
		fakeTimer.SumReturns(7 * int64(time.Millisecond))
		tc.registry.Register("test-timer", fakeTimer)
		tc.exporter.exportMetrics(tc.registry)

		Expect(tc.transportMessages).To(Receive(ContainElement(
			&dataPoint{
				Name:      "test-timer.duration.sum",
				Type:      "gauge",
				Value:     7,
				Timestamp: 123,
				Unit:      "milliseconds",
			},
		)))
	})
})

type fakeTimeHelper struct {
	returnValue int64
}

func newFakeTimeHelper() *fakeTimeHelper {
	return &fakeTimeHelper{}
}

func (f *fakeTimeHelper) currentTimeInMillis() int64 {
	return f.returnValue
}

type fakeTransporter struct {
	messages chan []*dataPoint
}

func newFakeTransporter(messages chan []*dataPoint) *fakeTransporter {
	return &fakeTransporter{
		messages: messages,
	}
}

func (f *fakeTransporter) send(data []*dataPoint) error {
	f.messages <- data
	return nil
}
