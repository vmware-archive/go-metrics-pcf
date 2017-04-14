package pcf

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
func (h *httpTransporter) send([]*dataPoint) error {
	return nil
}
