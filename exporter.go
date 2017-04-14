package pcf

import (
	"github.com/rcrowley/go-metrics"
	"log"
	"time"
)

type dataPoint struct {
	metricType string
	value      float64
}

type transporter interface {
	Send([]*dataPoint) error
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
	metricForwarderUrl := "http://example.com/v1/"
	apiToken := "test-token"
	instanceId := "d61e5f10-16a4-47fc-bdf9-f8a5c097cf7b"
	instanceIndex := 1

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
	exporter := newExporter(transport)

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
}

func newExporter(transport transporter) *exporter {
	return &exporter{
		transport: transport,
	}
}

func (e *exporter) exportMetrics(registry metrics.Registry) error {
	dataPoints := e.assembleDataPoints(registry)

	return e.transport.Send(dataPoints)
}

func (e *exporter) assembleDataPoints(registry metrics.Registry) []*dataPoint {
	// TODO convert `go-metrics metrics` to `pcf metrics`
	return make([]*dataPoint, 0)
}

type httpTransporter struct {
	url   string
	token string
}

func newHttpTransporter(url string, token string) *httpTransporter {
	return &httpTransporter{
		url:   url,
		token: token,
	}
}

// TODO: post to url with authorization header
func (h *httpTransporter) Send([]*dataPoint) error {
	return nil
}
