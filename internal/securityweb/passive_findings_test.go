package securityweb

import "testing"

func TestPassiveObservation_RequiresIDStatusTitleEvidenceLink(t *testing.T) {
	ok := NewPassiveObservation(PassiveObservation{
		ObservationID: "obs-1",
		Title:         "Content-Security-Policy",
		Status:        ObservationStatusMissing,
		EvidenceID:    "ev-1",
		SourceCheckID: PassiveCheckIDHeadersMVP,
	})
	if !ObservationHasMinimumFields(ok) {
		t.Fatalf("expected valid observation minimum fields")
	}
	bad := PassiveObservation{Title: "x", Status: ObservationStatusPresent, EvidenceID: "ev"}
	if ObservationHasMinimumFields(bad) {
		t.Fatalf("expected invalid observation without ID")
	}
}

func TestCandidateFinding_RequiresCoreFieldsAndLinksObservationIDs(t *testing.T) {
	f := NewCandidateFinding(CandidateFinding{
		CandidateID:           "cand-1",
		Title:                 "Missing CSP",
		Category:              FindingCategoryConfiguration,
		SeverityHint:          SeverityLow,
		ConfidenceHint:        ConfidenceHigh,
		EvidenceID:            "ev-1",
		RelatedObservationIDs: []string{"obs-csp"},
		SourceCheckID:         PassiveCheckIDHeadersMVP,
	})
	if !CandidateHasMinimumFields(f) {
		t.Fatalf("expected valid candidate finding")
	}
	if len(f.RelatedObservationIDs) == 0 {
		t.Fatalf("expected related observation ids")
	}
}

func TestSeverityAndConfidenceEnumsAligned(t *testing.T) {
	severities := []SeverityHint{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInformational}
	for _, s := range severities {
		if s == "" {
			t.Fatalf("severity enum value must not be empty")
		}
	}
	conf := []ConfidenceHint{ConfidenceHigh, ConfidenceMedium, ConfidenceLow}
	for _, c := range conf {
		if c == "" {
			t.Fatalf("confidence enum value must not be empty")
		}
	}
}
