package securityweb

import (
	"testing"
	"time"
)

func TestRobotsPresentWithoutSensitivePaths_NoFinding(t *testing.T) {
	res := EvaluatePassiveRobots(ExecutionResult{EvidenceID: "ev-r1", URL: "https://example.com/robots.txt", BodyPreviewRedacted: "User-agent: *\nDisallow: /tmp"})
	if !hasObsID(res, "robots_present") {
		t.Fatalf("expected robots_present observation")
	}
	if len(res.CandidateFindings) != 0 {
		t.Fatalf("expected no candidate findings")
	}
}

func TestRobotsMissing_NoFinding(t *testing.T) {
	res := EvaluatePassiveRobots(ExecutionResult{EvidenceID: "ev-r2", URL: "https://example.com/robots.txt", StatusCode: 404})
	if !hasObsID(res, "robots_missing") {
		t.Fatalf("expected robots_missing")
	}
	if len(res.CandidateFindings) != 0 {
		t.Fatalf("expected no findings")
	}
}

func TestRobotsSensitivePaths_GeneratesConservativeFinding(t *testing.T) {
	res := EvaluatePassiveRobots(ExecutionResult{EvidenceID: "ev-r3", URL: "https://example.com/robots.txt", BodyPreviewRedacted: "User-agent: *\nDisallow: /admin\nDisallow: /backup"})
	if !hasObsID(res, "robots_disallow_sensitive_paths") {
		t.Fatalf("expected sensitive paths observation")
	}
	if !hasFindingID(res, "robots-sensitive-paths-listed") {
		t.Fatalf("expected conservative finding")
	}
}

func TestRobotsSitemapReferenceObservation(t *testing.T) {
	res := EvaluatePassiveRobots(ExecutionResult{EvidenceID: "ev-r4", URL: "https://example.com/robots.txt", BodyPreviewRedacted: "Sitemap: https://example.com/sitemap.xml"})
	if !hasObsID(res, "robots_sitemap_reference") {
		t.Fatalf("expected sitemap reference observation")
	}
	assertEvidenceAndSource(t, res, PassiveCheckIDRobotsMVP)
}

func TestSitemapPresentInScope_NoFinding(t *testing.T) {
	xml := "<urlset><url><loc>https://example.com/a</loc></url></urlset>"
	res := EvaluatePassiveSitemap(ExecutionResult{EvidenceID: "ev-s1", URL: "https://example.com/sitemap.xml", BodyPreviewRedacted: xml}, []string{"example.com"})
	if !hasObsID(res, "sitemap_present") || !hasObsID(res, "sitemap_urls_detected") {
		t.Fatalf("expected sitemap present and urls detected")
	}
	if len(res.CandidateFindings) != 0 {
		t.Fatalf("expected no findings")
	}
}

func TestSitemapMissing_NoFinding(t *testing.T) {
	res := EvaluatePassiveSitemap(ExecutionResult{EvidenceID: "ev-s2", URL: "https://example.com/sitemap.xml", StatusCode: 404}, []string{"example.com"})
	if !hasObsID(res, "sitemap_missing") {
		t.Fatalf("expected sitemap_missing")
	}
	if len(res.CandidateFindings) != 0 {
		t.Fatalf("expected no findings")
	}
}

func TestSitemapOutOfScopeWithScopeHosts_Finding(t *testing.T) {
	xml := "<urlset><url><loc>https://evil.example.org/x</loc></url></urlset>"
	res := EvaluatePassiveSitemap(ExecutionResult{EvidenceID: "ev-s3", URL: "https://example.com/sitemap.xml", BodyPreviewRedacted: xml}, []string{"example.com"})
	if !hasObsID(res, "sitemap_out_of_scope_urls") {
		t.Fatalf("expected out of scope observation")
	}
	if !hasFindingID(res, "sitemap-out-of-scope-urls") {
		t.Fatalf("expected out of scope finding")
	}
}

func TestSitemapOutOfScopeEmptyScopeHosts_NeedsContextNoFinding(t *testing.T) {
	xml := "<urlset><url><loc>https://evil.example.org/x</loc></url></urlset>"
	res := EvaluatePassiveSitemap(ExecutionResult{EvidenceID: "ev-s4", URL: "https://example.com/sitemap.xml", BodyPreviewRedacted: xml}, nil)
	if !hasObsID(res, "sitemap_needs_context") {
		t.Fatalf("expected needs_context observation")
	}
	if hasFindingID(res, "sitemap-out-of-scope-urls") {
		t.Fatalf("expected no out-of-scope finding without scope hosts")
	}
}

func TestSitemapSensitivePathFinding(t *testing.T) {
	xml := "<urlset><url><loc>https://example.com/admin</loc></url></urlset>"
	res := EvaluatePassiveSitemap(ExecutionResult{EvidenceID: "ev-s5", URL: "https://example.com/sitemap.xml", BodyPreviewRedacted: xml}, []string{"example.com"})
	if !hasObsID(res, "sitemap_sensitive_paths") {
		t.Fatalf("expected sensitive path observation")
	}
	if !hasFindingID(res, "sitemap-sensitive-paths") {
		t.Fatalf("expected sensitive path finding")
	}
}

func TestSitemapInvalidXML_NoPanicNeedsContext(t *testing.T) {
	res := EvaluatePassiveSitemap(ExecutionResult{EvidenceID: "ev-s6", URL: "https://example.com/sitemap.xml", BodyPreviewRedacted: "<urlset><url>"}, []string{"example.com"})
	if !hasObsID(res, "sitemap_needs_context") {
		t.Fatalf("expected needs_context on invalid xml")
	}
}

func TestSecurityTxtPresentWithContact_NoFinding(t *testing.T) {
	body := "Contact: mailto:security@example.com\nExpires: 2099-01-01T00:00:00Z"
	res := EvaluatePassiveSecurityTxt(ExecutionResult{EvidenceID: "ev-t1", URL: "https://example.com/.well-known/security.txt", BodyPreviewRedacted: body}, func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
	if !hasObsID(res, "securitytxt_contact_present") {
		t.Fatalf("expected contact observation")
	}
	if len(res.CandidateFindings) != 0 {
		t.Fatalf("expected no finding")
	}
}

func TestSecurityTxtMissing_InformationalNoFinding(t *testing.T) {
	res := EvaluatePassiveSecurityTxt(ExecutionResult{EvidenceID: "ev-t2", URL: "https://example.com/.well-known/security.txt", StatusCode: 404}, nil)
	if !hasObsID(res, "securitytxt_missing") {
		t.Fatalf("expected securitytxt_missing")
	}
	if len(res.CandidateFindings) != 0 {
		t.Fatalf("expected no findings")
	}
}

func TestSecurityTxtMissingOrStaleExpires_ConservativeFinding(t *testing.T) {
	body := "Contact: mailto:security@example.com\nExpires: 2020-01-01T00:00:00Z"
	res := EvaluatePassiveSecurityTxt(ExecutionResult{EvidenceID: "ev-t3", URL: "https://example.com/.well-known/security.txt", BodyPreviewRedacted: body}, func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
	if !hasObsID(res, "securitytxt_expires_missing_or_stale") {
		t.Fatalf("expected stale expires observation")
	}
	if !hasFindingID(res, "securitytxt-expires-stale") {
		t.Fatalf("expected stale expires finding")
	}
}

func TestSecurityTxtInvalidExpires_NeedsContext(t *testing.T) {
	body := "Contact: mailto:security@example.com\nExpires: not-a-date"
	res := EvaluatePassiveSecurityTxt(ExecutionResult{EvidenceID: "ev-t4", URL: "https://example.com/.well-known/security.txt", BodyPreviewRedacted: body}, nil)
	if !hasObsID(res, "securitytxt_needs_context") {
		t.Fatalf("expected needs_context for invalid expires")
	}
}

func TestAllResultsIncludeEvidenceAndRelatedObsAndSourceCheckID(t *testing.T) {
	res := EvaluatePassiveRobots(ExecutionResult{EvidenceID: "ev-all", URL: "https://example.com/robots.txt", BodyPreviewRedacted: "Disallow: /admin"})
	if res.EvidenceID != "ev-all" {
		t.Fatalf("result must include evidence")
	}
	for _, f := range res.CandidateFindings {
		if len(f.RelatedObservationIDs) == 0 {
			t.Fatalf("candidate finding must link related observations")
		}
	}
	assertEvidenceAndSource(t, res, PassiveCheckIDRobotsMVP)
}

func assertEvidenceAndSource(t *testing.T, res PassiveCheckResult, checkID PassiveCheckID) {
	t.Helper()
	for _, o := range res.Observations {
		if o.EvidenceID == "" {
			t.Fatalf("observation must include evidence id")
		}
		if o.SourceCheckID != checkID {
			t.Fatalf("unexpected source check id in observation")
		}
	}
	for _, f := range res.CandidateFindings {
		if f.EvidenceID == "" {
			t.Fatalf("finding must include evidence id")
		}
		if f.SourceCheckID != checkID {
			t.Fatalf("unexpected source check id in finding")
		}
	}
}

func hasObsID(res PassiveCheckResult, id string) bool {
	for _, o := range res.Observations {
		if o.ObservationID == id {
			return true
		}
	}
	return false
}

func hasFindingID(res PassiveCheckResult, id string) bool {
	for _, f := range res.CandidateFindings {
		if f.CandidateID == id {
			return true
		}
	}
	return false
}
