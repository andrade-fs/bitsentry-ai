package securitywebhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bitsentry-ai/internal/securityweb"
)

var (
	ErrTimeoutRequired = errors.New("timeout required")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrTransport       = errors.New("transport error")
)

type Transport struct {
	client       *http.Client
	maxBodyBytes int64
}

func New(timeout time.Duration, maxBodyBytes int64) (*Transport, error) {
	if timeout <= 0 {
		return nil, ErrTimeoutRequired
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = 4096
	}

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Transport{client: client, maxBodyBytes: maxBodyBytes}, nil
}

func (t *Transport) Execute(req securityweb.PlannedRequest) (securityweb.FakeTransportResponse, error) {
	if t == nil || t.client == nil {
		return securityweb.FakeTransportResponse{}, fmt.Errorf("%w: client not initialized", ErrTransport)
	}
	if strings.TrimSpace(req.URL) == "" {
		return securityweb.FakeTransportResponse{}, fmt.Errorf("%w: missing url", ErrInvalidRequest)
	}

	method := string(req.Method)
	if method == "" {
		method = string(securityweb.MethodGET)
	}

	hreq, err := http.NewRequest(method, req.URL, nil)
	if err != nil {
		return securityweb.FakeTransportResponse{}, fmt.Errorf("%w: malformed request", ErrInvalidRequest)
	}
	for k, v := range req.Headers {
		hreq.Header.Set(k, v)
	}

	hresp, err := t.client.Do(hreq)
	if err != nil {
		return securityweb.FakeTransportResponse{}, normalizeDoError(err)
	}
	defer hresp.Body.Close()

	body, truncated, err := readBodyPreview(hresp.Body, t.maxBodyBytes)
	if err != nil {
		return securityweb.FakeTransportResponse{}, fmt.Errorf("%w: failed reading response body", ErrTransport)
	}

	return securityweb.FakeTransportResponse{
		StatusCode:       hresp.StatusCode,
		FinalURL:         finalURL(hresp),
		Headers:          flattenHeaders(hresp.Header),
		Body:             body,
		BodyTruncated:    truncated,
		RedirectObserved: isRedirect(hresp.StatusCode),
		RedirectLocation: strings.TrimSpace(hresp.Header.Get("Location")),
	}, nil
}

func normalizeDoError(err error) error {
	if err == nil {
		return nil
	}
	if ue, ok := err.(*url.Error); ok {
		if ue.Timeout() {
			return fmt.Errorf("%w: timeout", ErrTransport)
		}
		return fmt.Errorf("%w: request failed", ErrTransport)
	}
	return fmt.Errorf("%w: request failed", ErrTransport)
}

func readBodyPreview(r io.Reader, max int64) (string, bool, error) {
	if r == nil {
		return "", false, nil
	}
	if max <= 0 {
		max = 1
	}

	limited := io.LimitReader(r, max+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return "", false, err
	}
	if int64(len(b)) > max {
		return string(b[:max]), true, nil
	}
	return string(b), false, nil
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, values := range h {
		out[k] = strings.Join(values, ",")
	}
	return out
}

func finalURL(resp *http.Response) string {
	if resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return ""
	}
	return resp.Request.URL.String()
}

func isRedirect(status int) bool {
	return status >= 300 && status < 400
}
