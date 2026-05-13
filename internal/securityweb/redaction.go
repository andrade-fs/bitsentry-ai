package securityweb

import (
	"net/url"
	"strings"
)

type DefaultRedactor struct{}

func (r DefaultRedactor) RedactHeaders(headers map[string]string) map[string]string {
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		lk := strings.ToLower(k)
		if lk == "authorization" || lk == "cookie" || lk == "set-cookie" || lk == "x-api-key" || strings.HasPrefix(strings.ToLower(v), "bearer ") {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = v
	}
	return out
}

func (r DefaultRedactor) RedactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	sensitive := map[string]struct{}{
		"token": {}, "access_token": {}, "api_key": {}, "password": {}, "secret": {},
	}
	for k, vals := range q {
		if _, ok := sensitive[strings.ToLower(k)]; ok {
			for i := range vals {
				vals[i] = "[REDACTED]"
			}
			q[k] = vals
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
