package ecsmetadata

import (
	"io"
	"net/http"
	"strings"
)

type MockWrapper struct {
	Mock func(path string) (statusCode int, body string, err error)
}

func (m *MockWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.RequestURI()
	code, body, err := m.Mock(path)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Status:     http.StatusText(code),
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}
