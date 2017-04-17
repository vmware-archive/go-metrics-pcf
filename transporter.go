package pcf

import (
	"net/http"
	"bytes"
	"encoding/json"
	"fmt"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpTransporter struct {
	url   string
	token string
	client HttpClient
}

func newHttpTransporter(client HttpClient, url string, token string) *httpTransporter {
	return &httpTransporter{
		client: client,
		url:   url,
		token: token,
	}
}

func (h *httpTransporter) send(points []*dataPoint) error {
	body, err := json.Marshal(&points)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", h.token)
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
