package securityweb

import (
	"fmt"
	"strings"
)

type FakeTransportResponse struct {
	StatusCode       int
	FinalURL         string
	Headers          map[string]string
	Body             string
	RedirectObserved bool
	RedirectLocation string
}

type FakeTransport struct {
	byRequestRef map[string]FakeTransportResponse
	byMethodURL  map[string]FakeTransportResponse
}

func NewFakeTransport() *FakeTransport {
	return &FakeTransport{
		byRequestRef: map[string]FakeTransportResponse{},
		byMethodURL:  map[string]FakeTransportResponse{},
	}
}

func (t *FakeTransport) Register(requestRef string, method RequestMethod, rawURL string, resp FakeTransportResponse) {
	if t == nil {
		return
	}
	if strings.TrimSpace(requestRef) != "" {
		t.byRequestRef[requestRef] = resp
		return
	}
	t.byMethodURL[transportKey(method, rawURL)] = resp
}

func (t *FakeTransport) Execute(req PlannedRequest) (FakeTransportResponse, error) {
	if t == nil {
		return FakeTransportResponse{}, fmt.Errorf("fake transport is required")
	}
	if strings.TrimSpace(req.RequestRef) != "" {
		if resp, ok := t.byRequestRef[req.RequestRef]; ok {
			return resp, nil
		}
	}
	if resp, ok := t.byMethodURL[transportKey(req.Method, req.URL)]; ok {
		return resp, nil
	}
	return FakeTransportResponse{}, fmt.Errorf("no fake transport response for request")
}

func transportKey(method RequestMethod, rawURL string) string {
	return string(method) + " " + rawURL
}
