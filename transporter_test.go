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

		transporter := newHttpTransporter(fakeHttpClient, "http://example.com/metrics", "test-token")

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

		Expect(req.Header.Get("Authorization")).To(Equal("test-token"))
		Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

		var result []dataPoint
		bytes, _ := ioutil.ReadAll(req.Body)
		json.Unmarshal(bytes, &result)

		Expect(result).To(ConsistOf(dataPoint{
			Name:      "test-counter",
			Type:      "COUNTER",
			Value:     123,
			Timestamp: 872828732,
			Unit:      "counts",
		}))
	})

	It("returns an error if the status is non 2xx", func() {
		fakeHttpClient := newFakeHttpClient()
		fakeHttpClient.returnCode = 500

		transporter := newHttpTransporter(fakeHttpClient, "http://example.com/metrics", "test-token")

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
