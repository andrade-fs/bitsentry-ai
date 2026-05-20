package securityweb

import (
	"sort"
	"strings"
)

func BuildRiskHypothesesFromSurfaceMap(m SurfaceMap) RiskHypothesisSet {
	set := RiskHypothesisSet{
		SetID:       "risk-hypotheses-static-mvp",
		FromMapID:   m.MapID,
		EvidenceIDs: append([]string{}, m.EvidenceIDs...),
		Limitations: []string{"Hypotheses are triage proposals from passive evidence only", "Priority is triage order, not confirmed vulnerability severity"},
	}

	byID := map[string]*RiskHypothesis{}
	add := func(h RiskHypothesis) {
		if strings.TrimSpace(h.HypothesisID) == "" {
			return
		}
		existing, ok := byID[h.HypothesisID]
		if !ok {
			h.SourceCandidateAreaIDs = appendUnique(h.SourceCandidateAreaIDs)
			h.SourceSignalIDs = appendUnique(h.SourceSignalIDs)
			h.EvidenceIDs = appendUnique(h.EvidenceIDs)
			h.AffectedURLs = appendUnique(h.AffectedURLs)
			h.AffectedPaths = appendUnique(h.AffectedPaths)
			h.SuggestedExpertSkills = dedupSkills(h.SuggestedExpertSkills)
			h.SuggestedNextChecks = dedupNextChecks(h.SuggestedNextChecks)
			copyH := h
			byID[h.HypothesisID] = &copyH
			return
		}
		existing.SourceCandidateAreaIDs = appendUnique(existing.SourceCandidateAreaIDs, h.SourceCandidateAreaIDs...)
		existing.SourceSignalIDs = appendUnique(existing.SourceSignalIDs, h.SourceSignalIDs...)
		existing.EvidenceIDs = appendUnique(existing.EvidenceIDs, h.EvidenceIDs...)
		existing.AffectedURLs = appendUnique(existing.AffectedURLs, h.AffectedURLs...)
		existing.AffectedPaths = appendUnique(existing.AffectedPaths, h.AffectedPaths...)
		existing.SuggestedExpertSkills = dedupSkills(append(existing.SuggestedExpertSkills, h.SuggestedExpertSkills...))
		existing.SuggestedNextChecks = dedupNextChecks(append(existing.SuggestedNextChecks, h.SuggestedNextChecks...))
		existing.Limitations = appendUnique(existing.Limitations, h.Limitations...)
		if riskPriorityRank(h.Priority) > riskPriorityRank(existing.Priority) {
			existing.Priority = h.Priority
		}
		if confidenceRank(h.ConfidenceHint) > confidenceRank(existing.ConfidenceHint) {
			existing.ConfidenceHint = h.ConfidenceHint
		}
		if existing.Status == RiskHypothesisProposed && h.Status == RiskHypothesisNeedsMoreEvidence {
			existing.Status = h.Status
		}
	}

	for _, area := range m.CandidateAreas {
		evidence := append([]string{}, area.EvidenceIDs...)
		status := RiskHypothesisProposed
		priority := RiskPriorityLow
		confidence := area.ConfidenceHint
		if confidence == "" {
			confidence = ConfidenceMedium
		}
		sourceSignals := signalIDsByArea(area.AreaID, m.Signals)
		affectedURLs, affectedPaths := affectedByArea(area.AreaID, m)

		switch area.AreaID {
		case "security_headers":
			if confidence == ConfidenceHigh {
				priority = RiskPriorityMedium
			}
			add(RiskHypothesis{
				HypothesisID:           "rh-security-headers-gap",
				Title:                  "Security headers hardening gap",
				Category:               FindingCategoryConfiguration,
				Priority:               priority,
				ConfidenceHint:         confidence,
				SourceCandidateAreaIDs: []string{area.AreaID},
				SourceSignalIDs:        sourceSignals,
				EvidenceIDs:            evidence,
				AffectedURLs:           affectedURLs,
				AffectedPaths:          affectedPaths,
				SuggestedExpertSkills:  []SuggestedExpertSkill{SkillHeadersSecurityReview},
				SuggestedNextChecks:    []SuggestedNextCheck{NextHeadersHardeningReview},
				Reason:                 "Header posture signals suggest hardening opportunities",
				Limitations:            []string{"No active validation was executed"},
				Status:                 status,
			})
		case "exposure_from_robots":
			skills := []SuggestedExpertSkill{SkillExposureReview, SkillAccessControlReview}
			skills = append(skills, sensitivitySkills(affectedPaths)...)
			next := []SuggestedNextCheck{NextRobotsExposureReview, NextAccessControlReviewDryRun}
			next = append(next, sensitivityNextChecks(affectedPaths)...)
			add(RiskHypothesis{
				HypothesisID:           "rh-exposure-robots",
				Title:                  "Potentially sensitive paths advertised in robots.txt",
				Category:               FindingCategoryExposure,
				Priority:               RiskPriorityLow,
				ConfidenceHint:         confidence,
				SourceCandidateAreaIDs: []string{area.AreaID},
				SourceSignalIDs:        sourceSignals,
				EvidenceIDs:            evidence,
				AffectedURLs:           affectedURLs,
				AffectedPaths:          affectedPaths,
				SuggestedExpertSkills:  dedupSkills(skills),
				SuggestedNextChecks:    dedupNextChecks(next),
				Reason:                 "Passive robots signals indicate exposed route hints",
				Limitations:            []string{"robots hints are not direct exploit evidence"},
				Status:                 status,
			})
		case "exposure_from_sitemap":
			skills := []SuggestedExpertSkill{SkillExposureReview}
			skills = append(skills, sensitivitySkills(affectedPaths)...)
			next := []SuggestedNextCheck{NextSitemapExposureReview}
			next = append(next, sensitivityNextChecks(affectedPaths)...)
			add(RiskHypothesis{
				HypothesisID:           "rh-exposure-sitemap",
				Title:                  "Potentially sensitive or unexpected URLs disclosed in sitemap",
				Category:               FindingCategoryExposure,
				Priority:               RiskPriorityLow,
				ConfidenceHint:         confidence,
				SourceCandidateAreaIDs: []string{area.AreaID},
				SourceSignalIDs:        sourceSignals,
				EvidenceIDs:            evidence,
				AffectedURLs:           affectedURLs,
				AffectedPaths:          affectedPaths,
				SuggestedExpertSkills:  dedupSkills(skills),
				SuggestedNextChecks:    dedupNextChecks(next),
				Reason:                 "Sitemap-derived passive exposure should be triaged",
				Limitations:            []string{"No URL interaction was performed"},
				Status:                 status,
			})
		case "security_contact":
			add(RiskHypothesis{
				HypothesisID:           "rh-security-contact-posture",
				Title:                  "Security contact posture signal",
				Category:               FindingCategoryInformational,
				Priority:               RiskPriorityLow,
				ConfidenceHint:         confidence,
				SourceCandidateAreaIDs: []string{area.AreaID},
				SourceSignalIDs:        sourceSignals,
				EvidenceIDs:            evidence,
				AffectedURLs:           affectedURLs,
				AffectedPaths:          affectedPaths,
				SuggestedExpertSkills:  []SuggestedExpertSkill{SkillSecurityTxtReview},
				SuggestedNextChecks:    []SuggestedNextCheck{NextSecurityTxtGovernance},
				Reason:                 "security.txt related metadata can guide disclosure readiness review",
				Limitations:            []string{"Informational posture; not a vulnerability claim"},
				Status:                 status,
			})
		case "unknown_scope_needs_context":
			add(RiskHypothesis{
				HypothesisID:           "rh-unknown-scope-context",
				Title:                  "Scope context required before deeper triage",
				Category:               FindingCategoryInformational,
				Priority:               RiskPriorityMedium,
				ConfidenceHint:         confidence,
				SourceCandidateAreaIDs: []string{area.AreaID},
				SourceSignalIDs:        sourceSignals,
				EvidenceIDs:            evidence,
				AffectedURLs:           affectedURLs,
				AffectedPaths:          affectedPaths,
				SuggestedExpertSkills:  []SuggestedExpertSkill{SkillExposureReview},
				SuggestedNextChecks:    []SuggestedNextCheck{NextScopeClarification},
				Reason:                 "Scope is missing or ambiguous for passive evidence interpretation",
				Limitations:            []string{"Do not run active checks until scope is clarified"},
				Status:                 RiskHypothesisNeedsMoreEvidence,
			})
		case "out_of_scope_reference":
			add(RiskHypothesis{
				HypothesisID:           "rh-out-of-scope-reference",
				Title:                  "Out-of-scope reference requires boundary clarification",
				Category:               FindingCategoryExposure,
				Priority:               RiskPriorityMedium,
				ConfidenceHint:         confidence,
				SourceCandidateAreaIDs: []string{area.AreaID},
				SourceSignalIDs:        sourceSignals,
				EvidenceIDs:            evidence,
				AffectedURLs:           affectedURLs,
				AffectedPaths:          affectedPaths,
				SuggestedExpertSkills:  []SuggestedExpertSkill{SkillExposureReview},
				SuggestedNextChecks:    []SuggestedNextCheck{NextScopeClarification},
				Reason:                 "Potential out-of-scope assets observed from passive evidence",
				Limitations:            []string{"No interaction allowed with out-of-scope assets", "Clarify program ownership before any testing"},
				Status:                 RiskHypothesisNeedsMoreEvidence,
			})
		}
	}

	for _, h := range byID {
		set.Hypotheses = append(set.Hypotheses, *h)
	}
	sort.Slice(set.Hypotheses, func(i, j int) bool { return set.Hypotheses[i].HypothesisID < set.Hypotheses[j].HypothesisID })
	set.EvidenceIDs = appendUnique(set.EvidenceIDs)
	set.Limitations = appendUnique(set.Limitations)
	return set
}

func signalIDsByArea(areaID string, signals []SurfaceSignal) []string {
	ids := []string{}
	for _, s := range signals {
		if strings.Contains(strings.ToLower(s.SignalID), strings.ToLower(areaIDKeyword(areaID))) {
			ids = append(ids, s.SignalID)
		}
	}
	return appendUnique(ids)
}

func areaIDKeyword(areaID string) string {
	switch areaID {
	case "security_headers":
		return "headers"
	case "exposure_from_robots":
		return "robots"
	case "exposure_from_sitemap":
		return "sitemap"
	case "security_contact":
		return "securitytxt"
	case "unknown_scope_needs_context":
		return "needs_context"
	case "out_of_scope_reference":
		return "out_of_scope"
	default:
		return areaID
	}
}

func affectedByArea(areaID string, m SurfaceMap) ([]string, []string) {
	urls := []string{}
	paths := []string{}
	for _, u := range m.URLs {
		switch areaID {
		case "security_headers":
			if u.Source == SurfaceSourceHeaders {
				urls = append(urls, u.URL)
				paths = append(paths, u.Path)
			}
		case "exposure_from_robots":
			if u.Source == SurfaceSourceRobots {
				urls = append(urls, u.URL)
				paths = append(paths, u.Path)
			}
		case "exposure_from_sitemap", "out_of_scope_reference", "unknown_scope_needs_context":
			if u.Source == SurfaceSourceSitemap {
				urls = append(urls, u.URL)
				paths = append(paths, u.Path)
			}
		case "security_contact":
			if u.Source == SurfaceSourceSecurityTxt {
				urls = append(urls, u.URL)
				paths = append(paths, u.Path)
			}
		}
	}
	for _, p := range m.Paths {
		if areaID == "exposure_from_robots" && p.Source == SurfaceSourceRobots {
			paths = append(paths, p.Path)
		}
		if areaID == "exposure_from_sitemap" && p.Source == SurfaceSourceSitemap {
			paths = append(paths, p.Path)
		}
	}
	return appendUnique(urls), appendUnique(paths)
}

func sensitivitySkills(paths []string) []SuggestedExpertSkill {
	for _, p := range paths {
		s := ClassifyPathSensitivity(p)
		if s == SensitivityAdmin || s == SensitivityAuth || s == SensitivityPrivate || s == SensitivityStaging {
			return []SuggestedExpertSkill{SkillAuthSecurityReview, SkillAccessControlReview}
		}
	}
	return nil
}

func sensitivityNextChecks(paths []string) []SuggestedNextCheck {
	for _, p := range paths {
		s := ClassifyPathSensitivity(p)
		if s == SensitivityAdmin || s == SensitivityAuth || s == SensitivityPrivate || s == SensitivityStaging {
			return []SuggestedNextCheck{NextAuthBoundaryReview, NextAccessControlReviewDryRun}
		}
	}
	return nil
}

func dedupSkills(in []SuggestedExpertSkill) []SuggestedExpertSkill {
	set := map[SuggestedExpertSkill]struct{}{}
	for _, v := range in {
		if strings.TrimSpace(string(v)) != "" {
			set[v] = struct{}{}
		}
	}
	out := make([]SuggestedExpertSkill, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func dedupNextChecks(in []SuggestedNextCheck) []SuggestedNextCheck {
	set := map[SuggestedNextCheck]struct{}{}
	for _, v := range in {
		if strings.TrimSpace(string(v)) != "" {
			set[v] = struct{}{}
		}
	}
	out := make([]SuggestedNextCheck, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func riskPriorityRank(p RiskPriority) int {
	switch p {
	case RiskPriorityHigh:
		return 3
	case RiskPriorityMedium:
		return 2
	default:
		return 1
	}
}

func confidenceRank(c ConfidenceHint) int {
	switch c {
	case ConfidenceHigh:
		return 3
	case ConfidenceMedium:
		return 2
	default:
		return 1
	}
}
