package securityweb

import "testing"

func TestBuildSurfaceMap_FromSitemapURLs(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDSitemapMVP,
		EvidenceID: "ev-1",
		Observations: []PassiveObservation{{ObservationID: "sitemap_urls_detected", Title: "sitemap_urls_detected", Status: ObservationStatusPresent, EvidenceID: "ev-1", ObservedValue: "https://example.com/a,https://example.com/b", SourceCheckID: PassiveCheckIDSitemapMVP}},
	}}
	m := BuildSurfaceMap(nil, checks, []string{"example.com"})
	if len(m.URLs) < 2 {
		t.Fatalf("expected urls from sitemap observations")
	}
}

func TestBuildSurfaceMap_FromRobotsDisallowPaths(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDRobotsMVP,
		EvidenceID: "ev-2",
		Observations: []PassiveObservation{{ObservationID: "robots_disallow_sensitive_paths", Title: "robots_disallow_sensitive_paths", Status: ObservationStatusPresent, EvidenceID: "ev-2", ObservedValue: "/admin,/backup", SourceCheckID: PassiveCheckIDRobotsMVP}},
	}}
	m := BuildSurfaceMap(nil, checks, []string{"example.com"})
	if !hasSurfacePath(m, "/admin") || !hasSurfacePath(m, "/backup") {
		t.Fatalf("expected paths from robots disallow")
	}
}

func TestBuildSurfaceMap_SecurityContactAreaAndSignal(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDSecurityTxtMVP,
		EvidenceID: "ev-3",
		Observations: []PassiveObservation{{ObservationID: "securitytxt_contact_present", Title: "securitytxt_contact_present", Status: ObservationStatusPresent, EvidenceID: "ev-3", SourceCheckID: PassiveCheckIDSecurityTxtMVP}},
	}}
	m := BuildSurfaceMap(nil, checks, nil)
	if !hasCandidateArea(m, "security_contact") {
		t.Fatalf("expected security_contact area")
	}
	if !hasSignal(m, string(PassiveCheckIDSecurityTxtMVP)+":securitytxt_contact_present") {
		t.Fatalf("expected contact signal")
	}
}

func TestBuildSurfaceMap_MissingFilesCreateSignalsNotFindings(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDRobotsMVP,
		EvidenceID: "ev-4",
		Observations: []PassiveObservation{{ObservationID: "robots_missing", Title: "robots_missing", Status: ObservationStatusMissing, EvidenceID: "ev-4", SourceCheckID: PassiveCheckIDRobotsMVP}},
	}}
	m := BuildSurfaceMap(nil, checks, nil)
	if len(m.Signals) == 0 {
		t.Fatalf("expected signals")
	}
}

func TestBuildSurfaceMap_OutOfScopeWhenScopeProvided(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDSitemapMVP,
		EvidenceID: "ev-5",
		Observations: []PassiveObservation{{ObservationID: "sitemap_out_of_scope_urls", Title: "sitemap_out_of_scope_urls", Status: ObservationStatusPresent, EvidenceID: "ev-5", ObservedValue: "https://evil.example.org/a", SourceCheckID: PassiveCheckIDSitemapMVP}},
	}}
	m := BuildSurfaceMap(nil, checks, []string{"example.com"})
	if !hasCandidateArea(m, "out_of_scope_reference") {
		t.Fatalf("expected out_of_scope_reference area")
	}
}

func TestBuildSurfaceMap_EmptyScopeUnknownNotOutOfScope(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDSitemapMVP,
		EvidenceID: "ev-6",
		Observations: []PassiveObservation{{ObservationID: "sitemap_needs_context", Title: "sitemap_needs_context", Status: ObservationStatusNeedsContext, EvidenceID: "ev-6", SourceCheckID: PassiveCheckIDSitemapMVP}},
	}}
	m := BuildSurfaceMap(nil, checks, nil)
	if !hasCandidateArea(m, "unknown_scope_needs_context") {
		t.Fatalf("expected unknown_scope_needs_context area")
	}
	if hasCandidateArea(m, "out_of_scope_reference") {
		t.Fatalf("did not expect out_of_scope_reference without scope hosts")
	}
}

func TestBuildSurfaceMap_SensitivePathAreas(t *testing.T) {
	checks := []PassiveCheckResult{
		{CheckID: PassiveCheckIDRobotsMVP, EvidenceID: "ev-r", CandidateFindings: []CandidateFinding{{CandidateID: "r1", EvidenceID: "ev-r", SourceCheckID: PassiveCheckIDRobotsMVP}}},
		{CheckID: PassiveCheckIDSitemapMVP, EvidenceID: "ev-s", CandidateFindings: []CandidateFinding{{CandidateID: "s1", EvidenceID: "ev-s", SourceCheckID: PassiveCheckIDSitemapMVP}}},
	}
	m := BuildSurfaceMap(nil, checks, []string{"example.com"})
	if !hasCandidateArea(m, "exposure_from_robots") || !hasCandidateArea(m, "exposure_from_sitemap") {
		t.Fatalf("expected exposure candidate areas")
	}
}

func TestBuildSurfaceMap_DeduplicatesURLsAndPaths(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDSitemapMVP,
		EvidenceID: "ev-7",
		Observations: []PassiveObservation{{ObservationID: "sitemap_urls_detected", Title: "sitemap_urls_detected", Status: ObservationStatusPresent, EvidenceID: "ev-7", ObservedValue: "https://example.com/a,https://example.com/a", SourceCheckID: PassiveCheckIDSitemapMVP}},
	}}
	m := BuildSurfaceMap(nil, checks, []string{"example.com"})
	count := 0
	for _, u := range m.URLs {
		if u.Path == "/a" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected deduplicated url/path entries")
	}
}

func TestBuildSurfaceMap_PreservesEvidenceAndSignalTrace(t *testing.T) {
	checks := []PassiveCheckResult{{
		CheckID:    PassiveCheckIDHeadersMVP,
		EvidenceID: "ev-8",
		CandidateFindings: []CandidateFinding{{CandidateID: "hf-1", EvidenceID: "ev-8", SourceCheckID: PassiveCheckIDHeadersMVP, RelatedObservationIDs: []string{"obs-1"}}},
	}}
	m := BuildSurfaceMap(nil, checks, []string{"example.com"})
	if !contains(m.EvidenceIDs, "ev-8") {
		t.Fatalf("expected evidence id preserved")
	}
	id := string(PassiveCheckIDHeadersMVP) + ":hf-1"
	if !hasSignal(m, id) {
		t.Fatalf("expected finding-based signal")
	}
	for _, s := range m.Signals {
		if s.SignalID == id && len(s.RelatedObservationIDs) != 1 {
			t.Fatalf("expected related observation ids preserved")
		}
	}
}

func TestBuildSurfaceMap_DeterministicMapID(t *testing.T) {
	m := BuildSurfaceMap(nil, nil, nil)
	if m.MapID != "surface-map-static-mvp" {
		t.Fatalf("expected deterministic map id")
	}
}

func hasSurfacePath(m SurfaceMap, p string) bool {
	for _, v := range m.Paths {
		if v.Path == p {
			return true
		}
	}
	return false
}

func hasCandidateArea(m SurfaceMap, id string) bool {
	for _, a := range m.CandidateAreas {
		if a.AreaID == id {
			return true
		}
	}
	return false
}

func hasSignal(m SurfaceMap, id string) bool {
	for _, s := range m.Signals {
		if s.SignalID == id {
			return true
		}
	}
	return false
}

func contains(arr []string, v string) bool {
	for _, x := range arr {
		if x == v {
			return true
		}
	}
	return false
}
