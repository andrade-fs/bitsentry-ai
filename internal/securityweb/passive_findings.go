package securityweb

import "strings"

const (
	DefaultAffectedComponentHeaders   = "http_response_headers"
	DefaultAffectedComponentDiscovery = "passive_discovery_files"
)

func NewPassiveObservation(obs PassiveObservation) PassiveObservation {
	obs.ObservationID = strings.TrimSpace(obs.ObservationID)
	obs.Title = strings.TrimSpace(obs.Title)
	obs.EvidenceID = strings.TrimSpace(obs.EvidenceID)
	obs.AffectedURL = strings.TrimSpace(obs.AffectedURL)
	obs.AffectedComponent = strings.TrimSpace(obs.AffectedComponent)
	if obs.AffectedComponent == "" {
		obs.AffectedComponent = DefaultAffectedComponentHeaders
	}
	return obs
}

func NewCandidateFinding(f CandidateFinding) CandidateFinding {
	f.CandidateID = strings.TrimSpace(f.CandidateID)
	f.Title = strings.TrimSpace(f.Title)
	f.EvidenceID = strings.TrimSpace(f.EvidenceID)
	f.AffectedURL = strings.TrimSpace(f.AffectedURL)
	f.AffectedComponent = strings.TrimSpace(f.AffectedComponent)
	if f.AffectedComponent == "" {
		f.AffectedComponent = DefaultAffectedComponentHeaders
	}
	return f
}

func ObservationHasMinimumFields(obs PassiveObservation) bool {
	return strings.TrimSpace(obs.ObservationID) != "" && strings.TrimSpace(obs.Title) != "" && strings.TrimSpace(obs.EvidenceID) != "" && strings.TrimSpace(string(obs.Status)) != ""
}

func CandidateHasMinimumFields(f CandidateFinding) bool {
	if strings.TrimSpace(f.CandidateID) == "" || strings.TrimSpace(f.EvidenceID) == "" {
		return false
	}
	if strings.TrimSpace(string(f.SeverityHint)) == "" || strings.TrimSpace(string(f.ConfidenceHint)) == "" {
		return false
	}
	return true
}
