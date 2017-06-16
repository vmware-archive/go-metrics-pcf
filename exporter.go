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
	"log"
	"net/http"
	"time"

	"crypto/tls"

	"github.com/rcrowley/go-metrics"
)

const defaultCfMetricsServiceName = "cf-metrics"

type dataPoint struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
	Unit      string  `json:"unit"`
}

type transporter interface {
	send([]*dataPoint) error
}

type timeHelper interface {
	currentTimeInMillis() int64
}

// Options is used when starting an exporter.
type Options struct {
	Frequency           time.Duration
	InstanceId          string
	InstanceIndex       string
	Token               string
	Url                 string
	AppGuid             string
	TimeUnit            time.Duration
	ServiceName         string
	SkipSSLVerification bool
}

func (o *Options) fillDefaults() {
	if o.Token == "" {
		apiToken, err := getToken(o.ServiceName)
		if err != nil {
			log.Printf("Could not get apiToken: %s", err.Error())
			return
		}

		o.Token = apiToken
	}

	if o.Url == "" {
		apiUrl, err := getUrl(o.ServiceName)
		if err != nil {
			log.Printf("Could not get Url: %s", err.Error())
			return
		}

		o.Url = apiUrl
	}

	if o.AppGuid == "" {
		appGuid, err := getAppGuid()
		if err != nil {
			log.Printf("Could not get app guid: %s", err.Error())
			return
		}

		o.AppGuid = appGuid
	}
}

// ExporterOption is used to configure an exporter.
type ExporterOption func(*Options)

// WithFrequency sets the frequency. The default is a minute.
func WithFrequency(f time.Duration) ExporterOption {
	return func(o *Options) {
		o.Frequency = f
	}
}

// WithInstanceId sets the instance ID. The default is read from the
// environment variable INSTANCE_GUID
func WithInstanceId(guid string) ExporterOption {
	return func(o *Options) {
		o.InstanceId = guid
	}
}

// WithInstanceIndex sets the instance index. The default is read from the
// environment variable INSTANCE_INDEX.
func WithInstanceIndex(id string) ExporterOption {
	return func(o *Options) {
		o.InstanceId = id
	}
}

// WithToken sets the token.
func WithToken(token string) ExporterOption {
	return func(o *Options) {
		o.Token = token
	}
}

// WithURL sets the URL.
func WithURL(URL string) ExporterOption {
	return func(o *Options) {
		o.Url = URL
	}
}

// WithAppGuid sets the AppGuid.
func WithAppGuid(guid string) ExporterOption {
	return func(o *Options) {
		o.AppGuid = guid
	}
}

// WithTimeUnit sets the TimeUnit.
func WithTimeUnit(u time.Duration) ExporterOption {
	return func(o *Options) {
		o.TimeUnit = u
	}
}

// WithServiceName sets the ServiceName.
func WithServiceName(n string) ExporterOption {
	return func(o *Options) {
		o.ServiceName = n
	}
}

// WithSkipSSL sets the InsecureSkipVerify flag on the HTTP transport.
func WithSkipSSL(skip bool) ExporterOption {
	return func(o *Options) {
		o.SkipSSLVerification = skip
	}
}

// StartExporter starts a new exporter on the current go-routine and will
// never exit.
func StartExporter(registry metrics.Registry, opts ...ExporterOption) {
	options := &Options{
		Frequency:     time.Minute,
		InstanceIndex: getInstanceIndex(),
		InstanceId:    getInstanceGuid(),
		ServiceName:   defaultCfMetricsServiceName,
		TimeUnit:      time.Millisecond,
	}

	for _, o := range opts {
		o(options)
	}

	options.fillDefaults()

	StartExporterWithOptions(registry, options)
}

// StartExporterWithOptions starts a new exporter with provided options on
// the current go-routine and will never exit.
func StartExporterWithOptions(registry metrics.Registry, options *Options) {
	options.fillDefaults()

	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: options.SkipSSLVerification},
	}
	client := &http.Client{Transport: httpTransport}

	timer := time.NewTimer(options.Frequency)
	transport := newHttpTransporter(client, options)
	exporter := newExporter(transport, &realTimeHelper{}, options.TimeUnit)

	for {
		<-timer.C
		timer.Reset(options.Frequency)

		err := exporter.exportMetrics(registry)
		if err != nil {
			log.Printf("Could not export metrics to PCF: %s", err.Error())
		}
	}
}

type exporter struct {
	transport  transporter
	timeHelper timeHelper
	timeUnit   time.Duration
}

func newExporter(transport transporter, timeHelper timeHelper, timeUnit time.Duration) *exporter {
	return &exporter{
		transport:  transport,
		timeHelper: timeHelper,
		timeUnit:   timeUnit,
	}
}

func (e *exporter) exportMetrics(registry metrics.Registry) error {
	dataPoints := e.assembleDataPoints(registry)

	return e.transport.send(dataPoints)
}

func (e *exporter) assembleDataPoints(registry metrics.Registry) []*dataPoint {
	data := make([]*dataPoint, 0)
	currentTime := e.timeHelper.currentTimeInMillis()

	registry.Each(func(name string, metric interface{}) {
		switch m := metric.(type) {
		case metrics.Counter:
			data = append(data, convertCounter(m.Snapshot(), name, currentTime))
		case metrics.Gauge:
			data = append(data, convertGauge(m.Snapshot(), name, currentTime))
		case metrics.GaugeFloat64:
			data = append(data, convertGaugeFloat64(m.Snapshot(), name, currentTime))
		case metrics.Meter:
			data = append(data, convertMeter(m.Snapshot(), name, currentTime)...)
		case metrics.Timer:
			data = append(data, convertTimer(m.Snapshot(), name, currentTime, e.timeUnit)...)
		case metrics.Histogram:
			data = append(data, convertHistogram(m.Snapshot(), name, currentTime)...)
		}
	})

	return data
}
