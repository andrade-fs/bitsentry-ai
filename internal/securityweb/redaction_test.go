package securityweb

import (
	"strings"
	"testing"
)

func TestRedactorRedactsHeadersAndSensitiveQueryParams(t *testing.T) {
	r := DefaultRedactor{}
	headers := map[string]string{
		"Authorization": "Bearer abc",
		"authorization": "Bearer xyz",
		"Cookie":        "sid=123",
		"set-cookie":    "sid=456",
		"X-API-Key":     "apikey",
		"X-Trace-ID":    "trace",
	}
	redactedHeaders := r.RedactHeaders(headers)
	for _, h := range []string{"Authorization", "authorization", "Cookie", "set-cookie", "X-API-Key"} {
		if redactedHeaders[h] != "[REDACTED]" {
			t.Fatalf("expected header %s to be redacted", h)
		}
	}
	if redactedHeaders["X-Trace-ID"] != "trace" {
		t.Fatalf("non-sensitive header must remain")
	}

	gotURL := r.RedactURL("https://example.com?a=1&token=t&access_token=a&api_key=k&password=p&secret=s&TOKEN=t2")
	for _, k := range []string{"token=%5BREDACTED%5D", "access_token=%5BREDACTED%5D", "api_key=%5BREDACTED%5D", "password=%5BREDACTED%5D", "secret=%5BREDACTED%5D", "TOKEN=%5BREDACTED%5D"} {
		if !strings.Contains(gotURL, k) {
			t.Fatalf("expected query param to be redacted: %s", k)
		}
	}
}
