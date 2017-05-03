package pcfmetrics

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

type Options struct {
	Frequency     time.Duration
	InstanceId    string
	InstanceIndex string
	Token         string
	Url           string
	AppGuid       string
	TimeUnit      time.Duration
	ServiceName   string
}

func StartExporter(registry metrics.Registry) {
	ExportWithOptions(registry, &Options{})
}

func ExportWithOptions(registry metrics.Registry, options *Options) {
	if options.ServiceName == "" {
		options.ServiceName = defaultCfMetricsServiceName
	}

	if options.Token == "" {
		apiToken, err := getToken(options.ServiceName)
		if err != nil {
			log.Printf("Could not get apiToken: %s", err.Error())
			return
		}

		options.Token = apiToken
	}

	if options.Url == "" {
		apiUrl, err := getUrl(options.ServiceName)
		if err != nil {
			log.Printf("Could not get Url: %s", err.Error())
			return
		}

		options.Url = apiUrl
	}

	if options.AppGuid == "" {
		appGuid, err := getAppGuid()
		if err != nil {
			log.Printf("Could not get app guid: %s", err.Error())
			return
		}

		options.AppGuid = appGuid
	}

	if options.InstanceIndex == "" {
		options.InstanceIndex = getInstanceIndex()
	}

	if options.InstanceId == "" {
		options.InstanceId = getInstanceGuid()
	}

	if int64(options.TimeUnit) == 0 {
		options.TimeUnit = time.Millisecond
	}

	if int64(options.Frequency) == 0 {
		options.Frequency = time.Minute
	}

	url := fmt.Sprintf("https://%s/apps/%s/instances/%s/%s", options.Url, options.AppGuid, options.InstanceId, options.InstanceIndex)

	timer := time.NewTimer(options.Frequency)
	transport := newHttpTransporter(http.DefaultClient, url, options.Token)
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
