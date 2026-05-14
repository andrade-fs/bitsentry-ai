package securityweb

import (
	"net/url"
	"strings"
)

type HeaderStatus string

const (
	HeaderStatusPresent       HeaderStatus = "present"
	HeaderStatusMissing       HeaderStatus = "missing"
	HeaderStatusWeak          HeaderStatus = "weak"
	HeaderStatusNotApplicable HeaderStatus = "not_applicable"
)

type SeverityHint string

const (
	SeverityCritical      SeverityHint = "Critical"
	SeverityHigh          SeverityHint = "High"
	SeverityMedium        SeverityHint = "Medium"
	SeverityLow           SeverityHint = "Low"
	SeverityInformational SeverityHint = "Informational"
)

type ConfidenceHint string

const (
	ConfidenceHigh   ConfidenceHint = "High"
	ConfidenceMedium ConfidenceHint = "Medium"
	ConfidenceLow    ConfidenceHint = "Low"
)

type HeaderCheckInput struct {
	ExecutionResult   ExecutionResult
	AllowGETFallback  bool
	RequestedMethod   RequestMethod
}

type HeaderObservation struct {
	ID            string
	Header        string
	Status        HeaderStatus
	Value         string
	SeverityHint  SeverityHint
	ConfidenceHint ConfidenceHint
	Notes         string
}

type CandidateFinding struct {
	ID             string
	Title          string
	Category       string
	SeverityHint   SeverityHint
	ConfidenceHint ConfidenceHint
	EvidenceID     string
	ObservationIDs []string
	Summary        string
}

type HeaderCheckResult struct {
	CheckID           string
	EvidenceID        string
	Observations      []HeaderObservation
	CandidateFindings []CandidateFinding
	Limitations       []string
}

func EvaluatePassiveHeaders(input HeaderCheckInput) HeaderCheckResult {
	r := input.ExecutionResult
	result := HeaderCheckResult{
		CheckID:    "passive_headers_mvp",
		EvidenceID: strings.TrimSpace(r.EvidenceID),
	}

	if result.EvidenceID == "" {
		result.Limitations = append(result.Limitations, "missing evidence_id in execution result")
	}

	headers := normalizeHeaders(r.HeadersRedacted)
	obs := make([]HeaderObservation, 0, 12)
	findings := make([]CandidateFinding, 0, 8)

	mkObs := func(id, header string, status HeaderStatus, value string, sev SeverityHint, conf ConfidenceHint, notes string) HeaderObservation {
		return HeaderObservation{ID: id, Header: header, Status: status, Value: value, SeverityHint: sev, ConfidenceHint: conf, Notes: notes}
	}
	addFinding := func(id, title, category string, sev SeverityHint, conf ConfidenceHint, observationIDs []string, summary string) {
		findings = append(findings, CandidateFinding{
			ID:             id,
			Title:          title,
			Category:       category,
			SeverityHint:   sev,
			ConfidenceHint: conf,
			EvidenceID:     result.EvidenceID,
			ObservationIDs: observationIDs,
			Summary:        summary,
		})
	}

	// CSP
	csp := strings.TrimSpace(headers["content-security-policy"])
	if csp == "" {
		o := mkObs("obs-csp", "Content-Security-Policy", HeaderStatusMissing, "", SeverityLow, ConfidenceHigh, "missing CSP can increase XSS/clickjacking exposure depending on app context")
		obs = append(obs, o)
		addFinding("hf-csp-missing", "Missing Content-Security-Policy", "Configuration", SeverityLow, ConfidenceHigh, []string{o.ID}, "CSP is missing; impact requires endpoint/content context")
	} else {
		obs = append(obs, mkObs("obs-csp", "Content-Security-Policy", HeaderStatusPresent, csp, SeverityInformational, ConfidenceHigh, "header present"))
	}

	// HSTS applicability based on final URL or request URL
	urlForScheme := firstNonEmpty(strings.TrimSpace(r.FinalURL), strings.TrimSpace(r.URL))
	scheme := detectScheme(urlForScheme)
	hsts := strings.TrimSpace(headers["strict-transport-security"])
	if scheme == "https" {
		if hsts == "" {
			o := mkObs("obs-hsts", "Strict-Transport-Security", HeaderStatusMissing, "", SeverityMedium, ConfidenceHigh, "missing HSTS on HTTPS response")
			obs = append(obs, o)
			addFinding("hf-hsts-missing", "Missing HSTS on HTTPS", "Configuration", SeverityMedium, ConfidenceHigh, []string{o.ID}, "HSTS missing on HTTPS response")
		} else {
			obs = append(obs, mkObs("obs-hsts", "Strict-Transport-Security", HeaderStatusPresent, hsts, SeverityInformational, ConfidenceHigh, "header present"))
		}
	} else {
		obs = append(obs, mkObs("obs-hsts", "Strict-Transport-Security", HeaderStatusNotApplicable, hsts, SeverityInformational, ConfidenceHigh, "HSTS applies to HTTPS responses"))
		result.Limitations = append(result.Limitations, "HSTS not applicable because response URL is not HTTPS")
	}

	// X-Content-Type-Options
	xcto := strings.TrimSpace(headers["x-content-type-options"])
	if xcto == "" {
		o := mkObs("obs-xcto", "X-Content-Type-Options", HeaderStatusMissing, "", SeverityLow, ConfidenceHigh, "expected value nosniff")
		obs = append(obs, o)
		addFinding("hf-xcto-missing", "Missing X-Content-Type-Options", "Configuration", SeverityLow, ConfidenceHigh, []string{o.ID}, "X-Content-Type-Options missing")
	} else if !strings.EqualFold(xcto, "nosniff") {
		o := mkObs("obs-xcto", "X-Content-Type-Options", HeaderStatusWeak, xcto, SeverityLow, ConfidenceHigh, "recommended value is nosniff")
		obs = append(obs, o)
		addFinding("hf-xcto-weak", "Weak X-Content-Type-Options", "Configuration", SeverityLow, ConfidenceHigh, []string{o.ID}, "X-Content-Type-Options present with non-recommended value")
	} else {
		obs = append(obs, mkObs("obs-xcto", "X-Content-Type-Options", HeaderStatusPresent, xcto, SeverityInformational, ConfidenceHigh, "header present"))
	}

	// Clickjacking: XFO or CSP frame-ancestors
	xfo := strings.TrimSpace(headers["x-frame-options"])
	hasFrameAncestors := hasCSPDirective(csp, "frame-ancestors")
	if xfo == "" && !hasFrameAncestors {
		o := mkObs("obs-clickjacking", "X-Frame-Options / CSP frame-ancestors", HeaderStatusMissing, "", SeverityLow, ConfidenceMedium, "neither X-Frame-Options nor frame-ancestors found")
		obs = append(obs, o)
		addFinding("hf-clickjacking-protection-missing", "Missing clickjacking protection header", "Configuration", SeverityLow, ConfidenceMedium, []string{o.ID}, "No X-Frame-Options and no CSP frame-ancestors directive")
	} else if xfo == "" && hasFrameAncestors {
		obs = append(obs, mkObs("obs-clickjacking", "X-Frame-Options / CSP frame-ancestors", HeaderStatusPresent, "covered by CSP frame-ancestors", SeverityInformational, ConfidenceHigh, "X-Frame-Options missing but covered by CSP"))
	} else {
		obs = append(obs, mkObs("obs-clickjacking", "X-Frame-Options / CSP frame-ancestors", HeaderStatusPresent, xfo, SeverityInformational, ConfidenceHigh, "protection present"))
	}

	// Referrer-Policy
	ref := strings.TrimSpace(headers["referrer-policy"])
	if ref == "" {
		o := mkObs("obs-referrer", "Referrer-Policy", HeaderStatusMissing, "", SeverityLow, ConfidenceHigh, "missing referrer policy")
		obs = append(obs, o)
		addFinding("hf-referrer-missing", "Missing Referrer-Policy", "Configuration", SeverityLow, ConfidenceHigh, []string{o.ID}, "Referrer-Policy missing")
	} else if isWeakReferrerPolicy(ref) {
		o := mkObs("obs-referrer", "Referrer-Policy", HeaderStatusWeak, ref, SeverityLow, ConfidenceHigh, "policy may leak referrer data")
		obs = append(obs, o)
		addFinding("hf-referrer-weak", "Weak Referrer-Policy", "Configuration", SeverityLow, ConfidenceHigh, []string{o.ID}, "Referrer-Policy is weak")
	} else {
		obs = append(obs, mkObs("obs-referrer", "Referrer-Policy", HeaderStatusPresent, ref, SeverityInformational, ConfidenceHigh, "header present"))
	}

	// Additional recommended headers (conservative by default)
	checkOptional := func(id, header string) {
		v := strings.TrimSpace(headers[strings.ToLower(header)])
		if v == "" {
			o := mkObs(id, header, HeaderStatusMissing, "", SeverityInformational, ConfidenceHigh, "header missing; applicability depends on architecture/context")
			obs = append(obs, o)
			result.Limitations = append(result.Limitations, header+" applicability may require endpoint/context validation")
			return
		}
		obs = append(obs, mkObs(id, header, HeaderStatusPresent, v, SeverityInformational, ConfidenceHigh, "header present"))
	}
	checkOptional("obs-permissions", "Permissions-Policy")
	checkOptional("obs-coop", "Cross-Origin-Opener-Policy")
	checkOptional("obs-corp", "Cross-Origin-Resource-Policy")
	checkOptional("obs-coep", "Cross-Origin-Embedder-Policy")

	if input.RequestedMethod == MethodHEAD && !input.AllowGETFallback {
		result.Limitations = append(result.Limitations, "GET fallback may be needed in a future approved request")
	}

	result.Observations = obs
	result.CandidateFindings = findings
	return result
}

func normalizeHeaders(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	return out
}

func hasCSPDirective(cspValue, directive string) bool {
	if strings.TrimSpace(cspValue) == "" {
		return false
	}
	parts := strings.Split(strings.ToLower(cspValue), ";")
	d := strings.ToLower(strings.TrimSpace(directive))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, d+" ") || p == d {
			return true
		}
	}
	return false
}

func isWeakReferrerPolicy(v string) bool {
	norm := strings.ToLower(strings.TrimSpace(v))
	return norm == "unsafe-url" || norm == "origin-when-cross-origin"
}

func detectScheme(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(u.Scheme))
}
