package pcf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rcrowley/go-metrics"
)

var _ = Describe("`go-metrics` exporter for PCF Metrics", func() {
	It("exports metrics counter metrics", func() {
		registry := metrics.NewRegistry()
		transportMessages := make(chan []*dataPoint, 100)
		exporter := newExporter(newFakeTransporter(transportMessages))

		counter := metrics.NewCounter()
		counter.Inc(3)
		counter.Inc(3)

		registry.Register("test-counter", counter)


		exporter.exportMetrics(registry)


		Expect(transportMessages).To(Receive(Equal([]*dataPoint{
			{
				metricType: "counter",
				value: 6,
			},
		})))
	})

	// TODO:
	//   gauge
	//   gauge float64
	//   meter
	//   histogram
	//   timer
	//   ewma
})

type fakeTransporter struct {
	messages chan []*dataPoint
}

func newFakeTransporter(messages chan []*dataPoint) *fakeTransporter {
	return &fakeTransporter{
		messages: messages,
	}
}

func (f *fakeTransporter) Send(data []*dataPoint) error {
	f.messages <- data
	return nil
}
