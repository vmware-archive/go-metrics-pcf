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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpTransporter struct {
	client HttpClient
	options *Options
}

func newHttpTransporter(client HttpClient, options *Options) *httpTransporter {
	return &httpTransporter{
		client: client,
		options: options,
	}
}

func (h *httpTransporter) send(points []*dataPoint) error {
	payload := metricForwarderPayload{
		Applications: []metricForwarderApplication{
			{
				Id: h.options.AppGuid,
				Instances: []metricForwarderInstance{
					{
						Id: h.options.InstanceId,
						Index: h.options.InstanceIndex,
						Metrics: points,
					},
				},
			},
		},
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, h.options.Url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", h.options.Token)
	req.Header.Add("Content-Type", "application/json")

	res, err := h.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("Received a non-2xx status code: %d", res.StatusCode)
	}

	return nil
}

type metricForwarderPayload struct {
	Applications []metricForwarderApplication `json:"applications"`
}

type metricForwarderApplication struct {
	Id string `json:"id"`
	Instances []metricForwarderInstance `json:"instances"`
}

type metricForwarderInstance struct {
	Id string `json:"id"`
	Index string `json:"index"`
	Metrics []*dataPoint `json:"metrics"`
}