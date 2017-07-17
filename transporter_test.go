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
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("httpTransporter", func() {
	It("sends POST requests to the metric forwarder", func() {
		fakeHttpClient := newFakeHttpClient()
		fakeHttpClient.returnCode = 200

		transporter := newHttpTransporter(fakeHttpClient, &Options{
			AppGuid:       "some-application-id",
			InstanceId:    "some-instance-id",
			InstanceIndex: "1",
			Token:         "test-token",
			Url:           "https://whereTheMetricsGo.com/somepath",
		})

		err := transporter.send([]*dataPoint{
			{
				Name:      "test-counter",
				Type:      "COUNTER",
				Value:     123,
				Timestamp: 872828732,
				Unit:      "counts",
			},
		})

		Expect(err).To(Not(HaveOccurred()))

		req := <-fakeHttpClient.requests

		Expect(req.URL.Scheme).To(Equal("https"))
		Expect(req.URL.Host).To(Equal("whereTheMetricsGo.com"))
		Expect(req.URL.Path).To(Equal("somepath"))
		Expect(req.Header.Get("Authorization")).To(Equal("test-token"))
		Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

		var result metricForwarderPayload
		bytes, _ := ioutil.ReadAll(req.Body)
		json.Unmarshal(bytes, &result)

		Expect(result).To(Equal(metricForwarderPayload{
			Applications: []metricForwarderApplication{
				{
					Id: "some-application-id",
					Instances: []metricForwarderInstance{
						{
							Id:    "some-instance-id",
							Index: "1",
							Metrics: []*dataPoint{
								{
									Name:      "test-counter",
									Type:      "COUNTER",
									Value:     123,
									Timestamp: 872828732,
									Unit:      "counts",
								},
							},
						},
					},
				},
			},
		}))
	})

	It("returns an error if the status is non 2xx", func() {
		fakeHttpClient := newFakeHttpClient()
		fakeHttpClient.returnCode = 500

		transporter := newHttpTransporter(fakeHttpClient, &Options{
			AppGuid:       "some-application-id",
			InstanceId:    "some-instance-id",
			InstanceIndex: "1",
			Token:         "test-token",
			Url:           "",
		})

		err := transporter.send([]*dataPoint{
			{
				Name:      "test-counter",
				Type:      "COUNTER",
				Value:     123,
				Timestamp: 872828732,
				Unit:      "counts",
			},
		})

		Expect(err).To(HaveOccurred())
	})
})

type fakeHttpClient struct {
	returnCode  int
	returnError error
	requests    chan *http.Request
}

func (f *fakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	f.requests <- req

	return &http.Response{
		StatusCode: f.returnCode,
	}, f.returnError
}

func newFakeHttpClient() *fakeHttpClient {
	return &fakeHttpClient{
		requests: make(chan *http.Request, 100),
	}
}
