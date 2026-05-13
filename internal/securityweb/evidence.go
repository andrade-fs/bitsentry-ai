package securityweb

import "fmt"

type DefaultEvidenceRecorder struct {
	redactor Redactor
}

func NewDefaultEvidenceRecorder(redactor Redactor) DefaultEvidenceRecorder {
	return DefaultEvidenceRecorder{redactor: redactor}
}

func (r DefaultEvidenceRecorder) Render(entry EvidenceEntry) string {
	return fmt.Sprintf(`## Evidence Entry
- Evidence ID: %s
- Session/Mode: %s
- Authorization Ref: %s
- Scope Ref: %s
- Planned Request Ref: %s
- Policy Decision: %s
- Redaction Applied: %t
- Linked Finding IDs: %v
- Notes / Assumptions: %s
`, entry.EvidenceID, entry.SessionMode, entry.AuthorizationRef, entry.ScopeRef, entry.PlannedRequestRef, entry.PolicyDecision, entry.RedactionApplied, entry.LinkedFindingIDs, entry.NotesAssumptions)
}

func (r DefaultEvidenceRecorder) Redact(entry EvidenceEntry) EvidenceEntry {
	if r.redactor == nil {
		r.redactor = DefaultRedactor{}
	}
	entry.RequestHeaders = r.redactor.RedactHeaders(entry.RequestHeaders)
	entry.RequestURL = r.redactor.RedactURL(entry.RequestURL)
	entry.RedactionApplied = true
	return entry
}

func (a *OfflineAdapter) RenderEvidenceTemplate(entry EvidenceEntry) string {
	if a.recorder == nil {
		a.recorder = NewDefaultEvidenceRecorder(DefaultRedactor{})
	}
	return a.recorder.Render(entry)
}

func (a *OfflineAdapter) RedactEvidence(entry EvidenceEntry) EvidenceEntry {
	if a.recorder == nil {
		a.recorder = NewDefaultEvidenceRecorder(DefaultRedactor{})
	}
	return a.recorder.Redact(entry)
}
