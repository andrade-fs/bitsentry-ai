package securityweb

import (
	"net/url"
	"strings"
)

type HeaderCheckInput struct {
	ExecutionResult  ExecutionResult
	AllowGETFallback bool
	RequestedMethod  RequestMethod
}

func EvaluatePassiveHeaders(input HeaderCheckInput) HeaderCheckResult {
	r := input.ExecutionResult
	evidenceID := strings.TrimSpace(r.EvidenceID)
	affectedURL := firstNonEmpty(strings.TrimSpace(r.FinalURL), strings.TrimSpace(r.URL))

	result := HeaderCheckResult{
		CheckID:    PassiveCheckIDHeadersMVP,
		EvidenceID: evidenceID,
	}

	if evidenceID == "" {
		result.Limitations = append(result.Limitations, "missing evidence_id in execution result")
	}

	headers := normalizeHeaders(r.HeadersRedacted)
	obs := make([]PassiveObservation, 0, 12)
	findings := make([]CandidateFinding, 0, 8)

	mkObs := func(id, title string, status ObservationStatus, value string, sev SeverityHint, conf ConfidenceHint, notes string) PassiveObservation {
		return NewPassiveObservation(PassiveObservation{
			ObservationID:     id,
			Title:             title,
			Status:            status,
			EvidenceID:        evidenceID,
			AffectedURL:       affectedURL,
			AffectedComponent: DefaultAffectedComponentHeaders,
			ObservedValue:     value,
			SeverityHint:      sev,
			ConfidenceHint:    conf,
			Notes:             notes,
			SourceCheckID:     PassiveCheckIDHeadersMVP,
		})
	}
	addFinding := func(id, title string, sev SeverityHint, conf ConfidenceHint, observationIDs []string, impact, likelihood, remediation, verification string, limitations []string) {
		findings = append(findings, NewCandidateFinding(CandidateFinding{
			CandidateID:           id,
			Title:                 title,
			Category:              FindingCategoryConfiguration,
			SeverityHint:          sev,
			ConfidenceHint:        conf,
			EvidenceID:            evidenceID,
			RelatedObservationIDs: observationIDs,
			AffectedURL:           affectedURL,
			AffectedComponent:     DefaultAffectedComponentHeaders,
			Impact:                impact,
			Likelihood:            likelihood,
			Remediation:           remediation,
			Verification:          verification,
			Limitations:           limitations,
			SourceCheckID:         PassiveCheckIDHeadersMVP,
		}))
	}

	csp := strings.TrimSpace(headers["content-security-policy"])
	if csp == "" {
		o := mkObs("obs-csp", "Content-Security-Policy", ObservationStatusMissing, "", SeverityLow, ConfidenceHigh, "missing CSP can increase XSS/clickjacking exposure depending on app context")
		obs = append(obs, o)
		addFinding("hf-csp-missing", "Missing Content-Security-Policy", SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "Increases attack surface for content injection controls", "Medium", "Add a context-appropriate CSP baseline policy", "Confirm CSP header is present and enforced on target response", []string{"Missing header is not direct exploit proof"})
	} else {
		obs = append(obs, mkObs("obs-csp", "Content-Security-Policy", ObservationStatusPresent, csp, SeverityInformational, ConfidenceHigh, "header present"))
	}

	urlForScheme := affectedURL
	scheme := detectScheme(urlForScheme)
	hsts := strings.TrimSpace(headers["strict-transport-security"])
	if scheme == "https" {
		if hsts == "" {
			o := mkObs("obs-hsts", "Strict-Transport-Security", ObservationStatusMissing, "", SeverityMedium, ConfidenceHigh, "missing HSTS on HTTPS response")
			obs = append(obs, o)
			addFinding("hf-hsts-missing", "Missing HSTS on HTTPS", SeverityMedium, ConfidenceHigh, []string{o.ObservationID}, "May allow protocol downgrade risk in some client paths", "Medium", "Set Strict-Transport-Security on HTTPS responses", "Re-check HTTPS response includes HSTS", nil)
		} else {
			obs = append(obs, mkObs("obs-hsts", "Strict-Transport-Security", ObservationStatusPresent, hsts, SeverityInformational, ConfidenceHigh, "header present"))
		}
	} else {
		obs = append(obs, mkObs("obs-hsts", "Strict-Transport-Security", ObservationStatusNotApplicable, hsts, SeverityInformational, ConfidenceHigh, "HSTS applies to HTTPS responses"))
		result.Limitations = append(result.Limitations, "HSTS not applicable because response URL is not HTTPS")
	}

	xcto := strings.TrimSpace(headers["x-content-type-options"])
	if xcto == "" {
		o := mkObs("obs-xcto", "X-Content-Type-Options", ObservationStatusMissing, "", SeverityLow, ConfidenceHigh, "expected value nosniff")
		obs = append(obs, o)
		addFinding("hf-xcto-missing", "Missing X-Content-Type-Options", SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "Can weaken MIME sniffing protections", "Low", "Set X-Content-Type-Options: nosniff", "Confirm nosniff is returned", nil)
	} else if !strings.EqualFold(xcto, "nosniff") {
		o := mkObs("obs-xcto", "X-Content-Type-Options", ObservationStatusWeak, xcto, SeverityLow, ConfidenceHigh, "recommended value is nosniff")
		obs = append(obs, o)
		addFinding("hf-xcto-weak", "Weak X-Content-Type-Options", SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "Protection may not be effective", "Low", "Use exact value nosniff", "Confirm header value is nosniff", nil)
	} else {
		obs = append(obs, mkObs("obs-xcto", "X-Content-Type-Options", ObservationStatusPresent, xcto, SeverityInformational, ConfidenceHigh, "header present"))
	}

	xfo := strings.TrimSpace(headers["x-frame-options"])
	hasFrameAncestors := hasCSPDirective(csp, "frame-ancestors")
	if xfo == "" && !hasFrameAncestors {
		o := mkObs("obs-clickjacking", "X-Frame-Options / CSP frame-ancestors", ObservationStatusMissing, "", SeverityLow, ConfidenceMedium, "neither X-Frame-Options nor frame-ancestors found")
		obs = append(obs, o)
		addFinding("hf-clickjacking-protection-missing", "Missing clickjacking protection header", SeverityLow, ConfidenceMedium, []string{o.ObservationID}, "May increase framing risk", "Low", "Add X-Frame-Options or CSP frame-ancestors", "Verify one framing control exists", []string{"No direct exploit demonstrated in passive check"})
	} else if xfo == "" && hasFrameAncestors {
		obs = append(obs, mkObs("obs-clickjacking", "X-Frame-Options / CSP frame-ancestors", ObservationStatusPresent, "covered by CSP frame-ancestors", SeverityInformational, ConfidenceHigh, "X-Frame-Options missing but covered by CSP"))
	} else {
		obs = append(obs, mkObs("obs-clickjacking", "X-Frame-Options / CSP frame-ancestors", ObservationStatusPresent, xfo, SeverityInformational, ConfidenceHigh, "protection present"))
	}

	ref := strings.TrimSpace(headers["referrer-policy"])
	if ref == "" {
		o := mkObs("obs-referrer", "Referrer-Policy", ObservationStatusMissing, "", SeverityLow, ConfidenceHigh, "missing referrer policy")
		obs = append(obs, o)
		addFinding("hf-referrer-missing", "Missing Referrer-Policy", SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "May leak referrer data", "Low", "Set strict-origin-when-cross-origin or stricter", "Confirm policy is returned in response", nil)
	} else if isWeakReferrerPolicy(ref) {
		o := mkObs("obs-referrer", "Referrer-Policy", ObservationStatusWeak, ref, SeverityLow, ConfidenceHigh, "policy may leak referrer data")
		obs = append(obs, o)
		addFinding("hf-referrer-weak", "Weak Referrer-Policy", SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "May disclose extra URL context", "Low", "Use stricter referrer policy", "Confirm stricter policy in response", nil)
	} else {
		obs = append(obs, mkObs("obs-referrer", "Referrer-Policy", ObservationStatusPresent, ref, SeverityInformational, ConfidenceHigh, "header present"))
	}

	checkOptional := func(id, header string) {
		v := strings.TrimSpace(headers[strings.ToLower(header)])
		if v == "" {
			o := mkObs(id, header, ObservationStatusNeedsContext, "", SeverityInformational, ConfidenceHigh, "header missing; applicability depends on architecture/context")
			obs = append(obs, o)
			result.Limitations = append(result.Limitations, header+" applicability may require endpoint/context validation")
			return
		}
		obs = append(obs, mkObs(id, header, ObservationStatusPresent, v, SeverityInformational, ConfidenceHigh, "header present"))
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
