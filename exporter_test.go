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

package pcfmetrics_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/Benjamintf1/Expanded-Unmarshalled-Matchers"

	metricFakes "github.com/pivotal-cf/go-metrics-pcf/go-metrics-pcffakes"
	"github.com/rcrowley/go-metrics"
	"github.com/pivotal-cf/go-metrics-pcf"
	"net/http/httptest"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"fmt"
)

type metricForwarderPayload struct {
	Applications []*application `json:"applications"`
}

type application struct {
	Id        string `json:"id"`
	Instances []*instance `json:"instances"`
}

type instance struct {
	Id      string `json:"id"`
	Index   string `json:"index"`
	Metrics []*metric `json:"metrics"`
}

type metric struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     float64 `json:"value"`
	Unit      string `json:"unit"`
	Timestamp *int64 `json:"timestamp,omitempty"`
}

var _ = Describe("`go-metrics` exporter for PCF Metrics", func() {
	type testContext struct {
		registry                   metrics.Registry
		fakeMetricsForwarderServer *httptest.Server
		requestBodies              chan []byte
		requests                   chan *http.Request
	}

	var setup = func(responseCode int) *testContext {
		tc := &testContext{
			registry:      metrics.NewRegistry(),
			requestBodies: make(chan []byte, 100),
			requests:      make(chan *http.Request, 100),
		}

		tc.fakeMetricsForwarderServer = httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					body, err := ioutil.ReadAll(req.Body)
					Expect(err).ToNot(HaveOccurred())

					tc.requestBodies <- body
					tc.requests <- req

					w.WriteHeader(responseCode)
				},
			),
		)

		return tc
	}

	var setupAndStart = func(responseCode int) *testContext {
		tc := setup(responseCode)

		go pcfmetrics.StartExporter(
			tc.registry,
			pcfmetrics.WithFrequency(100*time.Millisecond),
			pcfmetrics.WithToken("fake-token"),
			pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
			pcfmetrics.WithAppGuid("fake-app-guid"),
		)
		return tc
	}

	var teardown = func(tc *testContext) {
		//TODO stop the exporter
		tc.fakeMetricsForwarderServer.CloseClientConnections()
		tc.fakeMetricsForwarderServer.Close()
	}

	Describe("metric format", func() {
		It("exports metrics with the current timestamp", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

			counter := metrics.NewCounter()
			counter.Inc(6)
			tc.registry.Register("test-counter", counter)

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))

			var payloadObject metricForwarderPayload
			err := json.Unmarshal(payload, &payloadObject)
			Expect(err).ToNot(HaveOccurred())

			Expect(payloadObject.Applications).To(HaveLen(1))

			instances := payloadObject.Applications[0].Instances
			Expect(instances).To(HaveLen(1))

			metrics := instances[0].Metrics
			Expect(metrics).To(HaveLen(1))

			unixTime := time.Unix(0, *metrics[0].Timestamp*int64(time.Millisecond))
			Expect(unixTime).To(BeTemporally("~", time.Now().UTC(), time.Second))
		})

		It("exports counter metrics", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

			counter := metrics.NewCounter()
			counter.Inc(6)
			tc.registry.Register("test-counter", counter)

			expectedJson := metricsToJsonString([]*metric{
				{
					Name:  "test-counter",
					Type:  "counter",
					Value: 6,
					Unit:  "",
				},
			})

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson))
		})

		It("exports gauge metrics", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

			gauge := metrics.NewGauge()
			gauge.Update(17)
			tc.registry.Register("test-gauge", gauge)

			expectedJson := metricsToJsonString([]*metric{
				{
					Name:  "test-gauge",
					Type:  "gauge",
					Value: 17,
					Unit:  "",
				},
			})

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson))
		})

		It("exports gauge float64 metrics", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

			gauge := metrics.NewGaugeFloat64()
			gauge.Update(32.2)
			tc.registry.Register("test-gauge-float64", gauge)

			expectedJson := metricsToJsonString([]*metric{
				{
					Name:  "test-gauge-float64",
					Type:  "gauge",
					Value: 32.2,
					Unit:  "",
				},
			})

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson))
		})

		It("exports meter metrics", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

			fakeMeter := new(metricFakes.FakeMeter)
			fakeMeter.SnapshotReturns(fakeMeter)
			fakeMeter.CountReturns(1)
			fakeMeter.Rate1Returns(2)
			fakeMeter.Rate5Returns(3)
			fakeMeter.Rate15Returns(4)
			fakeMeter.RateMeanReturns(5)

			tc.registry.Register("test-fakeMeter", fakeMeter)

			expectedJson := metricsToJsonString([]*metric{
				{
					Name:  "test-fakeMeter.count",
					Type:  "counter",
					Value: 1,
					Unit:  "",
				},
				{
					Name:  "test-fakeMeter.rate.1-minute",
					Type:  "gauge",
					Value: 2,
					Unit:  "",
				},
				{
					Name:  "test-fakeMeter.rate.5-minute",
					Type:  "gauge",
					Value: 3,
					Unit:  "",
				},
				{
					Name:  "test-fakeMeter.rate.15-minute",
					Type:  "gauge",
					Value: 4,
					Unit:  "",
				},
				{
					Name:  "test-fakeMeter.rate.mean",
					Type:  "gauge",
					Value: 5,
					Unit:  "",
				},
			})

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson, WithUnorderedListKeys("metrics")))
		})

		It("exports histogram metrics", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

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

			expectedJson := metricsToJsonString([]*metric{
				{
					Name:  "test-histogram.count",
					Type:  "counter",
					Value: 1,
					Unit:  "",
				},
				{
					Name:  "test-histogram.mean",
					Type:  "gauge",
					Value: 2,
					Unit:  "",
				},
				{
					Name:  "test-histogram.stddev",
					Type:  "gauge",
					Value: 3,
					Unit:  "",
				},
				{
					Name:  "test-histogram.sum",
					Type:  "gauge",
					Value: 4,
					Unit:  "",
				},
				{
					Name:  "test-histogram.variance",
					Type:  "gauge",
					Value: 5,
					Unit:  "",
				},
				{
					Name:  "test-histogram.max",
					Type:  "gauge",
					Value: 6,
					Unit:  "",
				},
				{
					Name:  "test-histogram.min",
					Type:  "gauge",
					Value: 7,
					Unit:  "",
				},
				{
					Name:  "test-histogram.75thPercentile",
					Type:  "gauge",
					Value: 8,
					Unit:  "",
				},
				{
					Name:  "test-histogram.95thPercentile",
					Type:  "gauge",
					Value: 9,
					Unit:  "",
				},
				{
					Name:  "test-histogram.98thPercentile",
					Type:  "gauge",
					Value: 10,
					Unit:  "",
				},
				{
					Name:  "test-histogram.99thPercentile",
					Type:  "gauge",
					Value: 11,
					Unit:  "",
				},
				{
					Name:  "test-histogram.999thPercentile",
					Type:  "gauge",
					Value: 12,
					Unit:  "",
				},
			})

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson, WithUnorderedListKeys("metrics")))
		})

		It("exports timer metrics", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

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
			expectedJson := metricsToJsonString([]*metric{
				{
					Name:  "test-timer.count",
					Type:  "counter",
					Value: 1,
					Unit:  "",
				},
				{
					Name:  "test-timer.rate.1-minute",
					Type:  "gauge",
					Value: 2,
					Unit:  "",
				},
				{
					Name:  "test-timer.rate.5-minute",
					Type:  "gauge",
					Value: 3,
					Unit:  "",
				},
				{
					Name:  "test-timer.rate.15-minute",
					Type:  "gauge",
					Value: 4,
					Unit:  "",
				},
				{
					Name:  "test-timer.rate.mean",
					Type:  "gauge",
					Value: 16,
					Unit:  "",
				},
				{
					Name:  "test-timer.duration.mean",
					Type:  "gauge",
					Value: 5,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.stddev",
					Type:  "gauge",
					Value: 6,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.sum",
					Type:  "gauge",
					Value: 7,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.variance",
					Type:  "gauge",
					Value: 8,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.max",
					Type:  "gauge",
					Value: 9,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.min",
					Type:  "gauge",
					Value: 10,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.75thPercentile",
					Type:  "gauge",
					Value: 11,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.95thPercentile",
					Type:  "gauge",
					Value: 12,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.98thPercentile",
					Type:  "gauge",
					Value: 13,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.99thPercentile",
					Type:  "gauge",
					Value: 14,
					Unit:  "milliseconds",
				},
				{
					Name:  "test-timer.duration.999thPercentile",
					Type:  "gauge",
					Value: 15,
					Unit:  "milliseconds",
				},
			})

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson, WithUnorderedListKeys("metrics")))
		})
	})

	It("continues to try and send metrics if metrics forwarder returns a bad status code", func() {
		tc := setupAndStart(http.StatusInternalServerError)
		defer teardown(tc)

		Eventually(tc.requestBodies).Should(HaveLen(2))
	})

	Describe("default options", func() {
		It("uses milliseconds as the default timer time unit", func() {
			tc := setupAndStart(http.StatusOK)
			defer teardown(tc)

			fakeTimer := new(metricFakes.FakeTimer)
			fakeTimer.SnapshotReturns(fakeTimer)
			fakeTimer.SumReturns(7 * int64(time.Millisecond))
			tc.registry.Register("test-timer", fakeTimer)

			expectedJson := metricsToJsonString([]*metric{
				{
					Name:  "test-timer.duration.sum",
					Type:  "gauge",
					Value: 7,
					Unit:  "milliseconds",
				},
			})

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson, WithUnorderedListKeys("metrics")))
		})

		It("sets the default metric rate to more than 5 seconds", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			Consistently(tc.requestBodies, 5).Should(HaveLen(0))
		})
	})

	Describe("setting options", func() {
		It("uses the correct url when using WithURL", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithFrequency(100*time.Millisecond),
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			Eventually(tc.requests).Should(Receive())
		})

		It("uses the correct access token when using WithToken", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithFrequency(100*time.Millisecond),
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			var req *http.Request
			Eventually(tc.requests).Should(Receive(&req))
			Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"fake-token"}))
		})

		It("sends the correct app guid when using WithAppGuid", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithFrequency(100*time.Millisecond),
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			app := wrapMetrics([]*metric{
				{
					Name: "test-counter",
					Type: "counter",
					Unit: "",
				},
			})

			expectedJson, err := json.Marshal(app)
			Expect(err).ToNot(HaveOccurred())

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson))
		})

		It("sends the correct instance id when using WithInstanceId", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithFrequency(100*time.Millisecond),
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
				pcfmetrics.WithInstanceId("fake-instance-id"),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			app := wrapMetrics([]*metric{
				{
					Name:  "test-counter",
					Type:  "counter",
					Unit:  "",
				},
			})

			app.Applications[0].Instances[0].Id = "fake-instance-id"
			expectedJson, err := json.Marshal(app)
			Expect(err).ToNot(HaveOccurred())

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson))
		})

		It("sends the correct instance index when using WithInstanceIndex", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithFrequency(100*time.Millisecond),
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
				pcfmetrics.WithInstanceIndex("fake-instance-index"),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			app := wrapMetrics([]*metric{
				{
					Name:  "test-counter",
					Type:  "counter",
					Unit:  "",
				},
			})

			app.Applications[0].Instances[0].Index = "fake-instance-index"
			expectedJson, err := json.Marshal(app)
			Expect(err).ToNot(HaveOccurred())

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson))
		})

		It("uses the correct time unit when using WithTimeUnit", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithFrequency(100*time.Millisecond),
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
				pcfmetrics.WithTimeUnit(time.Second),
			)

			fakeTimer := new(metricFakes.FakeTimer)
			fakeTimer.SnapshotReturns(fakeTimer)
			fakeTimer.MeanReturns(5 * float64(time.Millisecond))

			tc.registry.Register("test-timer", fakeTimer)

			app := wrapMetrics([]*metric{
				{
					Name:  "test-timer.duration.mean",
					Type:  "gauge",
					Value: .005,
					Unit:  "seconds",
				},
			})

			expectedJson, err := json.Marshal(app)
			Expect(err).ToNot(HaveOccurred())

			var payload []byte
			Eventually(tc.requestBodies).Should(Receive(&payload))
			Expect(string(payload)).To(ContainUnorderedJSON(expectedJson))
		})

		It("sends metrics at the correct frequency when using WithFrequency", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithToken("fake-token"),
				pcfmetrics.WithURL(tc.fakeMetricsForwarderServer.URL),
				pcfmetrics.WithAppGuid("fake-app-guid"),
				pcfmetrics.WithFrequency(100*time.Millisecond),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			Eventually(tc.requestBodies, 1).Should(HaveLen(5))
		})

		It("gets credentials from the correct service when using WithServiceName", func() {
			tc := setup(http.StatusOK)
			defer teardown(tc)

			vcapJson := fmt.Sprintf(`{
			  "fake-service-name": [
			   {
				"credentials": {
				 "access_key": "fake-access-key",
				 "endpoint": "%s"
				}
			   }
			  ]
			}`, tc.fakeMetricsForwarderServer.URL)

			os.Setenv("VCAP_SERVICES", vcapJson)

			go pcfmetrics.StartExporter(
				tc.registry,
				pcfmetrics.WithAppGuid("fake-app-guid"),
				pcfmetrics.WithFrequency(100*time.Millisecond),
				pcfmetrics.WithServiceName("fake-service-name"),
			)

			counter := metrics.NewCounter()
			tc.registry.Register("test-counter", counter)

			Eventually(tc.requestBodies, 1).Should(HaveLen(5))
		})
	})
})

func metricsToJsonString(metrics []*metric) string {
	bytes, err := json.Marshal(wrapMetrics(metrics))
	Expect(err).ToNot(HaveOccurred())

	return string(bytes)
}

func wrapMetrics(metrics []*metric) metricForwarderPayload {
	instances := []*instance{
		{
			Metrics: metrics,
		},
	}

	return metricForwarderPayload{
		Applications: []*application{
			{
				Id:        "fake-app-guid",
				Instances: instances,
			},
		},
	}
}
