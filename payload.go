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

type metricForwarderPayload struct {
	Applications []*metricForwarderApplication `json:"applications"`
}

func newMetricForwarderPayload(points []*dataPoint, options *Options) *metricForwarderPayload {
	return &metricForwarderPayload{
		Applications: []*metricForwarderApplication{
			newMetricForwarderApplication(points, options),
		},
	}
}

type metricForwarderApplication struct {
	Id        string `json:"id"`
	Instances []*metricForwarderInstance `json:"instances"`
}

func newMetricForwarderApplication(points []*dataPoint, options *Options) *metricForwarderApplication {
	return &metricForwarderApplication{
		Id: options.AppGuid,
		Instances: []*metricForwarderInstance{
			newMetricForwarderInstance(points, options),
		},
	}
}

type metricForwarderInstance struct {
	Id      string `json:"id"`
	Index   string `json:"index"`
	Metrics []*dataPoint `json:"metrics"`
}

func newMetricForwarderInstance(points []*dataPoint, options *Options) *metricForwarderInstance {
	return &metricForwarderInstance{
		Id:      options.InstanceId,
		Index:   options.InstanceIndex,
		Metrics: points,
	}
}