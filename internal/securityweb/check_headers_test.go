package securityweb

import "testing"

func TestPassiveHeaders_AllRecommendedPresent_NoCandidateFindings(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{
		ExecutionResult: ExecutionResult{
			EvidenceID: "WEB-EV-req-1-app-1",
			URL:        "https://example.com",
			FinalURL:   "https://example.com/home",
			HeadersRedacted: map[string]string{
				"Content-Security-Policy":      "default-src 'self'; frame-ancestors 'none'",
				"Strict-Transport-Security":    "max-age=31536000; includeSubDomains",
				"X-Content-Type-Options":       "nosniff",
				"X-Frame-Options":              "DENY",
				"Referrer-Policy":              "strict-origin-when-cross-origin",
				"Permissions-Policy":           "geolocation=()",
				"Cross-Origin-Opener-Policy":   "same-origin",
				"Cross-Origin-Resource-Policy": "same-origin",
				"Cross-Origin-Embedder-Policy": "require-corp",
			},
		},
		RequestedMethod:  MethodHEAD,
		AllowGETFallback: false,
	})
	if len(res.CandidateFindings) != 0 {
		t.Fatalf("expected no candidate findings, got %d", len(res.CandidateFindings))
	}
}

func TestHeaderCheckResult_UsesCommonPassiveModel(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev", URL: "https://example.com"}})
	if res.CheckID != PassiveCheckIDHeadersMVP {
		t.Fatalf("expected shared check id type")
	}
}

func TestPassiveHeaders_MissingCSP_GeneratesObservationAndCandidate(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-1", URL: "https://example.com", HeadersRedacted: map[string]string{}}})
	if !hasObservation(res, "Content-Security-Policy", ObservationStatusMissing) {
		t.Fatalf("expected missing CSP observation")
	}
	if !hasFinding(res, "hf-csp-missing") {
		t.Fatalf("expected CSP candidate finding")
	}
}

func TestPassiveHeaders_MissingHSTSOverHTTPS_GeneratesObservation(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-2", URL: "https://example.com", HeadersRedacted: map[string]string{}}})
	if !hasObservation(res, "Strict-Transport-Security", ObservationStatusMissing) {
		t.Fatalf("expected missing HSTS observation over https")
	}
}

func TestPassiveHeaders_HSTSOverHTTP_NotApplicableNoFinding(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-3", URL: "http://example.com", HeadersRedacted: map[string]string{}}})
	if !hasObservation(res, "Strict-Transport-Security", ObservationStatusNotApplicable) {
		t.Fatalf("expected HSTS not_applicable over http")
	}
	if hasFinding(res, "hf-hsts-missing") {
		t.Fatalf("did not expect HSTS finding over http")
	}
}

func TestPassiveHeaders_FrameAncestorsCoversMissingXFO(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-4", URL: "https://example.com", HeadersRedacted: map[string]string{"Content-Security-Policy": "default-src 'self'; frame-ancestors 'none'"}}})
	if hasFinding(res, "hf-clickjacking-protection-missing") {
		t.Fatalf("did not expect clickjacking finding when CSP frame-ancestors exists")
	}
}

func TestPassiveHeaders_MissingBothXFOAndFrameAncestors_FindingGenerated(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-5", URL: "https://example.com", HeadersRedacted: map[string]string{"Content-Security-Policy": "default-src 'self'"}}})
	if !hasFinding(res, "hf-clickjacking-protection-missing") {
		t.Fatalf("expected clickjacking candidate finding")
	}
}

func TestPassiveHeaders_PresentHeaderObservationNoFinding(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-p", URL: "https://example.com", HeadersRedacted: map[string]string{"X-Content-Type-Options": "nosniff"}}})
	if !hasObservation(res, "X-Content-Type-Options", ObservationStatusPresent) {
		t.Fatalf("expected present header observation")
	}
	if hasFinding(res, "hf-xcto-missing") || hasFinding(res, "hf-xcto-weak") {
		t.Fatalf("did not expect xcto finding when header is ok")
	}
}

func TestPassiveHeaders_WeakReferrerPolicyDetected(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-6", URL: "https://example.com", HeadersRedacted: map[string]string{"Referrer-Policy": "unsafe-url"}}})
	if !hasObservation(res, "Referrer-Policy", ObservationStatusWeak) {
		t.Fatalf("expected weak Referrer-Policy observation")
	}
}

func TestPassiveHeaders_MissingXCTODetected(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-7", URL: "https://example.com", HeadersRedacted: map[string]string{}}})
	if !hasObservation(res, "X-Content-Type-Options", ObservationStatusMissing) {
		t.Fatalf("expected missing X-Content-Type-Options observation")
	}
}

func TestPassiveHeaders_ResultLinksEvidenceIDAndSourceCheckID(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-8", URL: "https://example.com", HeadersRedacted: map[string]string{}}})
	if res.EvidenceID != "ev-8" {
		t.Fatalf("expected evidence id in result")
	}
	for _, o := range res.Observations {
		if o.EvidenceID != "ev-8" {
			t.Fatalf("expected observation evidence id")
		}
		if o.SourceCheckID != PassiveCheckIDHeadersMVP {
			t.Fatalf("expected source check id in observation")
		}
	}
	for _, f := range res.CandidateFindings {
		if f.EvidenceID != "ev-8" {
			t.Fatalf("expected finding evidence id")
		}
		if f.SourceCheckID != PassiveCheckIDHeadersMVP {
			t.Fatalf("expected source check id in finding")
		}
	}
}

func TestPassiveHeaders_SeverityAndConfidenceHintsPresent(t *testing.T) {
	res := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: ExecutionResult{EvidenceID: "ev-9", URL: "https://example.com", HeadersRedacted: map[string]string{}}})
	for _, o := range res.Observations {
		if o.SeverityHint == "" || o.ConfidenceHint == "" {
			t.Fatalf("expected observation severity/confidence hints")
		}
	}
	for _, f := range res.CandidateFindings {
		if f.SeverityHint == "" || f.ConfidenceHint == "" {
			t.Fatalf("expected finding severity/confidence hints")
		}
	}
}

func hasObservation(res HeaderCheckResult, title string, status ObservationStatus) bool {
	for _, o := range res.Observations {
		if o.Title == title && o.Status == status {
			return true
		}
	}
	return false
}

func hasFinding(res HeaderCheckResult, id string) bool {
	for _, f := range res.CandidateFindings {
		if f.CandidateID == id {
			return true
		}
	}
	return false
}
