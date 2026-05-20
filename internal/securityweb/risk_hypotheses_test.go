package securityweb

import (
	"strings"
	"testing"
)

func TestRiskHypotheses_SecurityHeadersRoutesToHeadersSkill(t *testing.T) {
	m := SurfaceMap{MapID: "surface-map-static-mvp", CandidateAreas: []SurfaceCandidateArea{{AreaID: "security_headers", Category: FindingCategoryConfiguration, EvidenceIDs: []string{"ev-1"}, ConfidenceHint: ConfidenceHigh}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-security-headers-gap")
	if h == nil || !hasSkill(h.SuggestedExpertSkills, SkillHeadersSecurityReview) {
		t.Fatalf("expected headers-security-review hypothesis")
	}
}

func TestRiskHypotheses_RobotsRoutesToExposureAndAccessControl(t *testing.T) {
	m := SurfaceMap{CandidateAreas: []SurfaceCandidateArea{{AreaID: "exposure_from_robots", EvidenceIDs: []string{"ev-2"}}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-exposure-robots")
	if h == nil || !hasSkill(h.SuggestedExpertSkills, SkillExposureReview) || !hasSkill(h.SuggestedExpertSkills, SkillAccessControlReview) {
		t.Fatalf("expected exposure and access-control skills")
	}
}

func TestRiskHypotheses_SitemapRoutesToExposure(t *testing.T) {
	m := SurfaceMap{CandidateAreas: []SurfaceCandidateArea{{AreaID: "exposure_from_sitemap", EvidenceIDs: []string{"ev-3"}}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-exposure-sitemap")
	if h == nil || !hasSkill(h.SuggestedExpertSkills, SkillExposureReview) {
		t.Fatalf("expected exposure-review skill")
	}
}

func TestRiskHypotheses_SecurityContactRoutesToSecurityTxtReview(t *testing.T) {
	m := SurfaceMap{CandidateAreas: []SurfaceCandidateArea{{AreaID: "security_contact", EvidenceIDs: []string{"ev-4"}}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-security-contact-posture")
	if h == nil || !hasSkill(h.SuggestedExpertSkills, SkillSecurityTxtReview) {
		t.Fatalf("expected securitytxt-review skill")
	}
}

func TestRiskHypotheses_SensitivePathsRouteToAuthAndAccessControl(t *testing.T) {
	m := SurfaceMap{
		CandidateAreas: []SurfaceCandidateArea{{AreaID: "exposure_from_robots", EvidenceIDs: []string{"ev-5"}}},
		Paths: []SurfacePath{{Path: "/admin/login", Source: SurfaceSourceRobots}, {Path: "/private", Source: SurfaceSourceRobots}},
	}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-exposure-robots")
	if h == nil || !hasSkill(h.SuggestedExpertSkills, SkillAuthSecurityReview) || !hasSkill(h.SuggestedExpertSkills, SkillAccessControlReview) {
		t.Fatalf("expected auth/access-control routing for sensitive paths")
	}
}

func TestRiskHypotheses_UnknownScopeNeedsMoreEvidence(t *testing.T) {
	m := SurfaceMap{CandidateAreas: []SurfaceCandidateArea{{AreaID: "unknown_scope_needs_context", EvidenceIDs: []string{"ev-6"}}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-unknown-scope-context")
	if h == nil || h.Status != RiskHypothesisNeedsMoreEvidence {
		t.Fatalf("expected needs_more_evidence status")
	}
}

func TestRiskHypotheses_OutOfScopeNeedsMoreEvidenceAndScopeClarification(t *testing.T) {
	m := SurfaceMap{CandidateAreas: []SurfaceCandidateArea{{AreaID: "out_of_scope_reference", EvidenceIDs: []string{"ev-7"}}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-out-of-scope-reference")
	if h == nil || h.Status != RiskHypothesisNeedsMoreEvidence {
		t.Fatalf("expected out_of_scope needs_more_evidence")
	}
	if !hasNextCheck(h.SuggestedNextChecks, NextScopeClarification) {
		t.Fatalf("expected scope clarification dry-run next check")
	}
}

func TestRiskHypotheses_DeduplicatesSimilarHypotheses(t *testing.T) {
	m := SurfaceMap{CandidateAreas: []SurfaceCandidateArea{{AreaID: "security_headers", EvidenceIDs: []string{"ev-1"}}, {AreaID: "security_headers", EvidenceIDs: []string{"ev-2"}}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	count := 0
	for _, h := range rs.Hypotheses {
		if h.HypothesisID == "rh-security-headers-gap" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected deduplicated security headers hypothesis")
	}
}

func TestRiskHypotheses_PreservesEvidenceAndSourceIDs(t *testing.T) {
	m := SurfaceMap{
		CandidateAreas: []SurfaceCandidateArea{{AreaID: "security_headers", EvidenceIDs: []string{"ev-8"}}},
		Signals: []SurfaceSignal{{SignalID: "passive_headers_mvp:obs-csp", SourceCheckID: PassiveCheckIDHeadersMVP, RelatedObservationIDs: []string{"obs-csp"}, EvidenceID: "ev-8"}},
		EvidenceIDs: []string{"ev-8"},
	}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	h := byHypothesisID(rs, "rh-security-headers-gap")
	if h == nil || !containsStr(h.EvidenceIDs, "ev-8") {
		t.Fatalf("expected evidence preservation")
	}
	if !containsStr(h.SourceCandidateAreaIDs, "security_headers") {
		t.Fatalf("expected source candidate area ids")
	}
}

func TestRiskHypotheses_SuggestedChecksDryRunOnly(t *testing.T) {
	m := SurfaceMap{CandidateAreas: []SurfaceCandidateArea{{AreaID: "security_headers"}, {AreaID: "exposure_from_robots"}, {AreaID: "security_contact"}}}
	rs := BuildRiskHypothesesFromSurfaceMap(m)
	for _, h := range rs.Hypotheses {
		for _, c := range h.SuggestedNextChecks {
			if !containsSubstring(string(c), "dry-run") {
				t.Fatalf("next check must be dry-run/planning only: %s", c)
			}
		}
	}
}

func TestRiskHypotheses_SetIDDeterministic(t *testing.T) {
	rs := BuildRiskHypothesesFromSurfaceMap(SurfaceMap{MapID: "surface-map-static-mvp"})
	if rs.SetID != "risk-hypotheses-static-mvp" {
		t.Fatalf("expected deterministic set id")
	}
}

func byHypothesisID(s RiskHypothesisSet, id string) *RiskHypothesis {
	for i := range s.Hypotheses {
		if s.Hypotheses[i].HypothesisID == id {
			return &s.Hypotheses[i]
		}
	}
	return nil
}

func hasSkill(skills []SuggestedExpertSkill, target SuggestedExpertSkill) bool {
	for _, s := range skills {
		if s == target {
			return true
		}
	}
	return false
}

func hasNextCheck(checks []SuggestedNextCheck, target SuggestedNextCheck) bool {
	for _, c := range checks {
		if c == target {
			return true
		}
	}
	return false
}

func containsStr(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func containsSubstring(v, part string) bool {
	return strings.Contains(v, part)
}
