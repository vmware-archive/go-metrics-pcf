package pcf

import (
	"github.com/rcrowley/go-metrics"
	"log"
	"time"
)

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
	frequency     time.Duration
	instanceId    string
	instanceIndex int
	token         string
	url           string
}

func Pcf(registry metrics.Registry) {
	// TODO: get values from environment
	metricForwarderUrl, apiToken, err := getCredentials()
	if err != nil {
		log.Printf("Could not get credentials: %s", err.Error())
		return
	}

	instanceIndex, err := getInstanceIndex()
	if err != nil {
		log.Printf("Could not get instance index: %s", err.Error())
		return
	}

	instanceId := getInstanceGuid()

	ExportWithOptions(registry, &Options{
		url:           metricForwarderUrl,
		token:         apiToken,
		frequency:     time.Minute,
		instanceId:    instanceId,
		instanceIndex: instanceIndex,
	})
}

func ExportWithOptions(registry metrics.Registry, options *Options) {
	timer := time.NewTimer(options.frequency)
	transport := newHttpTransporter(options.url, options.token)
	exporter := newExporter(transport, &realTimeHelper{})

	for {
		<-timer.C
		timer.Reset(options.frequency)

		err := exporter.exportMetrics(registry)
		if err != nil {
			log.Printf("Could not export metrics to PCF: %s", err.Error())
		}
	}
}

type exporter struct {
	transport transporter
	timeHelper timeHelper
}

func newExporter(transport transporter, timeHelper timeHelper) *exporter {
	return &exporter{
		transport: transport,
		timeHelper: timeHelper,
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
			data = append(data, convertTimer(m.Snapshot(), name, currentTime)...)
		case metrics.Histogram:
			data = append(data, convertHistogram(m.Snapshot(), name, currentTime)...)
		}
	})

	return data
}
