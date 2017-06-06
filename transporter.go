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

	url := fmt.Sprintf("https://%s/v1/metrics", h.options.Url)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
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