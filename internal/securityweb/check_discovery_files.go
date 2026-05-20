package securityweb

import (
	"encoding/xml"
	"net/url"
	"strings"
	"time"
)

type urlSet struct {
	URLs []urlEntry `xml:"url"`
}

type urlEntry struct {
	Loc string `xml:"loc"`
}

func EvaluatePassiveRobots(result ExecutionResult) PassiveCheckResult {
	evidenceID := strings.TrimSpace(result.EvidenceID)
	affectedURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), strings.TrimSpace(result.URL))
	body := strings.TrimSpace(result.BodyPreviewRedacted)

	out := PassiveCheckResult{CheckID: PassiveCheckIDRobotsMVP, EvidenceID: evidenceID}
	obs := []PassiveObservation{}
	findings := []CandidateFinding{}

	mkObs := func(id, title string, status ObservationStatus, value string, sev SeverityHint, conf ConfidenceHint, notes string) PassiveObservation {
		return NewPassiveObservation(PassiveObservation{ObservationID: id, Title: title, Status: status, EvidenceID: evidenceID, AffectedURL: affectedURL, AffectedComponent: DefaultAffectedComponentDiscovery, ObservedValue: value, SeverityHint: sev, ConfidenceHint: conf, Notes: notes, SourceCheckID: PassiveCheckIDRobotsMVP})
	}
	addFinding := func(id, title string, sev SeverityHint, conf ConfidenceHint, related []string, impact, likelihood, remediation, verification string) {
		findings = append(findings, NewCandidateFinding(CandidateFinding{CandidateID: id, Title: title, Category: FindingCategoryExposure, SeverityHint: sev, ConfidenceHint: conf, EvidenceID: evidenceID, RelatedObservationIDs: related, AffectedURL: affectedURL, AffectedComponent: DefaultAffectedComponentDiscovery, Impact: impact, Likelihood: likelihood, Remediation: remediation, Verification: verification, SourceCheckID: PassiveCheckIDRobotsMVP, Limitations: []string{"Passive signal only; no follow-up requests executed"}}))
	}

	if evidenceID == "" {
		out.Limitations = append(out.Limitations, "missing evidence_id in execution result")
	}
	if result.StatusCode == 404 || body == "" {
		obs = append(obs, mkObs("robots_missing", "robots_missing", ObservationStatusMissing, "", SeverityInformational, ConfidenceHigh, "robots.txt not present in provided response evidence"))
		out.Observations = obs
		out.CandidateFindings = findings
		return out
	}

	obs = append(obs, mkObs("robots_present", "robots_present", ObservationStatusPresent, "present", SeverityInformational, ConfidenceHigh, "robots.txt response present"))
	lc := strings.ToLower(body)
	if strings.Contains(lc, "sitemap:") {
		obs = append(obs, mkObs("robots_sitemap_reference", "robots_sitemap_reference", ObservationStatusPresent, "Sitemap directive found", SeverityInformational, ConfidenceHigh, "sitemap reference observed"))
	}
	if strings.Contains(lc, "allow: /") {
		obs = append(obs, mkObs("robots_unusual_allow_all", "robots_unusual_allow_all", ObservationStatusNeedsContext, "Allow: /", SeverityInformational, ConfidenceMedium, "allow-all may be normal depending on policy"))
	}

	sensitive := []string{"/admin", "/backup", "/private", "/staging"}
	matched := []string{}
	for _, p := range sensitive {
		if strings.Contains(lc, "disallow: "+p) || strings.Contains(lc, p) {
			matched = append(matched, p)
		}
	}
	if len(matched) > 0 {
		o := mkObs("robots_disallow_sensitive_paths", "robots_disallow_sensitive_paths", ObservationStatusPresent, strings.Join(matched, ","), SeverityLow, ConfidenceHigh, "sensitive-like paths listed in robots file")
		obs = append(obs, o)
		addFinding("robots-sensitive-paths-listed", "Sensitive-like paths listed in robots.txt", SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "May expose reconnaissance hints about administrative or internal routes", "Low", "Review whether listed paths should be externally discoverable", "Confirm path exposure policy with authorized owners")
	}

	out.Observations = obs
	out.CandidateFindings = findings
	return out
}

func EvaluatePassiveSitemap(result ExecutionResult, inScopeHosts []string) PassiveCheckResult {
	evidenceID := strings.TrimSpace(result.EvidenceID)
	affectedURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), strings.TrimSpace(result.URL))
	body := strings.TrimSpace(result.BodyPreviewRedacted)

	out := PassiveCheckResult{CheckID: PassiveCheckIDSitemapMVP, EvidenceID: evidenceID}
	obs := []PassiveObservation{}
	findings := []CandidateFinding{}

	mkObs := func(id, title string, status ObservationStatus, value string, sev SeverityHint, conf ConfidenceHint, notes string) PassiveObservation {
		return NewPassiveObservation(PassiveObservation{ObservationID: id, Title: title, Status: status, EvidenceID: evidenceID, AffectedURL: affectedURL, AffectedComponent: DefaultAffectedComponentDiscovery, ObservedValue: value, SeverityHint: sev, ConfidenceHint: conf, Notes: notes, SourceCheckID: PassiveCheckIDSitemapMVP})
	}
	addFinding := func(id, title string, category FindingCategory, sev SeverityHint, conf ConfidenceHint, related []string, impact, likelihood, remediation, verification string, limitations []string) {
		findings = append(findings, NewCandidateFinding(CandidateFinding{CandidateID: id, Title: title, Category: category, SeverityHint: sev, ConfidenceHint: conf, EvidenceID: evidenceID, RelatedObservationIDs: related, AffectedURL: affectedURL, AffectedComponent: DefaultAffectedComponentDiscovery, Impact: impact, Likelihood: likelihood, Remediation: remediation, Verification: verification, SourceCheckID: PassiveCheckIDSitemapMVP, Limitations: limitations}))
	}

	if evidenceID == "" {
		out.Limitations = append(out.Limitations, "missing evidence_id in execution result")
	}
	if result.StatusCode == 404 || body == "" {
		obs = append(obs, mkObs("sitemap_missing", "sitemap_missing", ObservationStatusMissing, "", SeverityInformational, ConfidenceHigh, "sitemap.xml not present in provided response evidence"))
		out.Observations = obs
		out.CandidateFindings = findings
		return out
	}

	obs = append(obs, mkObs("sitemap_present", "sitemap_present", ObservationStatusPresent, "present", SeverityInformational, ConfidenceHigh, "sitemap response present"))

	var parsed urlSet
	if err := xml.Unmarshal([]byte(body), &parsed); err != nil {
		obs = append(obs, mkObs("sitemap_needs_context", "sitemap_needs_context", ObservationStatusNeedsContext, "invalid XML", SeverityInformational, ConfidenceHigh, "sitemap XML could not be parsed from evidence preview"))
		out.Limitations = append(out.Limitations, "invalid sitemap XML in available evidence")
		out.Observations = obs
		out.CandidateFindings = findings
		return out
	}

	urls := []string{}
	for _, e := range parsed.URLs {
		loc := strings.TrimSpace(e.Loc)
		if loc != "" {
			urls = append(urls, loc)
		}
	}
	obs = append(obs, mkObs("sitemap_urls_detected", "sitemap_urls_detected", ObservationStatusPresent, strings.Join(urls, ","), SeverityInformational, ConfidenceHigh, "sitemap URLs detected from XML"))

	if len(inScopeHosts) == 0 {
		obs = append(obs, mkObs("sitemap_needs_context", "sitemap_needs_context", ObservationStatusNeedsContext, "scope list missing", SeverityInformational, ConfidenceHigh, "in-scope hosts not provided"))
	} else {
		outOfScope := []string{}
		for _, raw := range urls {
			h := parseHost(raw)
			if h != "" && !hostInScope(h, inScopeHosts) {
				outOfScope = append(outOfScope, raw)
			}
		}
		if len(outOfScope) > 0 {
			o := mkObs("sitemap_out_of_scope_urls", "sitemap_out_of_scope_urls", ObservationStatusPresent, strings.Join(outOfScope, ","), SeverityLow, ConfidenceHigh, "sitemap includes URLs outside provided scope")
			obs = append(obs, o)
			addFinding("sitemap-out-of-scope-urls", "Sitemap contains out-of-scope URLs", FindingCategoryExposure, SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "May indicate broader exposed surface than declared scope", "Low", "Review sitemap generation and scope boundaries", "Confirm URL ownership and intended exposure", []string{"No follow-up requests performed"})
		}
	}

	sensitiveMatches := []string{}
	for _, raw := range urls {
		lc := strings.ToLower(raw)
		if strings.Contains(lc, "/admin") || strings.Contains(lc, "/backup") || strings.Contains(lc, "/private") || strings.Contains(lc, "/staging") {
			sensitiveMatches = append(sensitiveMatches, raw)
		}
	}
	if len(sensitiveMatches) > 0 {
		o := mkObs("sitemap_sensitive_paths", "sitemap_sensitive_paths", ObservationStatusPresent, strings.Join(sensitiveMatches, ","), SeverityLow, ConfidenceMedium, "sitemap includes sensitive-like paths")
		obs = append(obs, o)
		addFinding("sitemap-sensitive-paths", "Sitemap lists sensitive-like paths", FindingCategoryExposure, SeverityLow, ConfidenceMedium, []string{o.ObservationID}, "Could aid passive reconnaissance", "Low", "Review whether listed paths should be discoverable", "Validate exposure policy with owners", []string{"Path naming alone does not confirm vulnerability"})
	}

	out.Observations = obs
	out.CandidateFindings = findings
	return out
}

func EvaluatePassiveSecurityTxt(result ExecutionResult, now func() time.Time) PassiveCheckResult {
	evidenceID := strings.TrimSpace(result.EvidenceID)
	affectedURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), strings.TrimSpace(result.URL))
	body := strings.TrimSpace(result.BodyPreviewRedacted)
	if now == nil {
		now = time.Now
	}

	out := PassiveCheckResult{CheckID: PassiveCheckIDSecurityTxtMVP, EvidenceID: evidenceID}
	obs := []PassiveObservation{}
	findings := []CandidateFinding{}

	mkObs := func(id, title string, status ObservationStatus, value string, sev SeverityHint, conf ConfidenceHint, notes string) PassiveObservation {
		return NewPassiveObservation(PassiveObservation{ObservationID: id, Title: title, Status: status, EvidenceID: evidenceID, AffectedURL: affectedURL, AffectedComponent: DefaultAffectedComponentDiscovery, ObservedValue: value, SeverityHint: sev, ConfidenceHint: conf, Notes: notes, SourceCheckID: PassiveCheckIDSecurityTxtMVP})
	}
	addFinding := func(id, title string, sev SeverityHint, conf ConfidenceHint, related []string, impact, likelihood, remediation, verification string, limitations []string) {
		findings = append(findings, NewCandidateFinding(CandidateFinding{CandidateID: id, Title: title, Category: FindingCategoryInformational, SeverityHint: sev, ConfidenceHint: conf, EvidenceID: evidenceID, RelatedObservationIDs: related, AffectedURL: affectedURL, AffectedComponent: DefaultAffectedComponentDiscovery, Impact: impact, Likelihood: likelihood, Remediation: remediation, Verification: verification, SourceCheckID: PassiveCheckIDSecurityTxtMVP, Limitations: limitations}))
	}

	if evidenceID == "" {
		out.Limitations = append(out.Limitations, "missing evidence_id in execution result")
	}
	if result.StatusCode == 404 || body == "" {
		obs = append(obs, mkObs("securitytxt_missing", "securitytxt_missing", ObservationStatusMissing, "", SeverityInformational, ConfidenceHigh, "security.txt not present in provided response evidence"))
		out.Observations = obs
		out.CandidateFindings = findings
		return out
	}

	obs = append(obs, mkObs("securitytxt_present", "securitytxt_present", ObservationStatusPresent, "present", SeverityInformational, ConfidenceHigh, "security.txt response present"))
	lc := strings.ToLower(body)
	if strings.Contains(lc, "contact:") {
		obs = append(obs, mkObs("securitytxt_contact_present", "securitytxt_contact_present", ObservationStatusPresent, "Contact field present", SeverityInformational, ConfidenceHigh, "contact information observed"))
	}
	if strings.Contains(lc, "policy:") {
		obs = append(obs, mkObs("securitytxt_policy_present", "securitytxt_policy_present", ObservationStatusPresent, "Policy field present", SeverityInformational, ConfidenceHigh, "policy reference observed"))
	}

	expiresVal := extractFieldValue(body, "Expires")
	if strings.TrimSpace(expiresVal) == "" {
		o := mkObs("securitytxt_expires_missing_or_stale", "securitytxt_expires_missing_or_stale", ObservationStatusMissing, "", SeverityInformational, ConfidenceHigh, "Expires field missing")
		obs = append(obs, o)
		addFinding("securitytxt-expires-missing", "security.txt missing Expires field", SeverityInformational, ConfidenceHigh, []string{o.ObservationID}, "Policy freshness window is unclear", "Low", "Add Expires field following RFC guidance", "Re-check security.txt includes future Expires date", nil)
	} else {
		ts, ok := parseSecurityTxtExpires(expiresVal)
		if !ok {
			obs = append(obs, mkObs("securitytxt_needs_context", "securitytxt_needs_context", ObservationStatusNeedsContext, expiresVal, SeverityInformational, ConfidenceHigh, "Expires field has invalid datetime format"))
			out.Limitations = append(out.Limitations, "invalid security.txt Expires format")
		} else if ts.Before(now()) {
			o := mkObs("securitytxt_expires_missing_or_stale", "securitytxt_expires_missing_or_stale", ObservationStatusWeak, expiresVal, SeverityLow, ConfidenceHigh, "Expires value is stale")
			obs = append(obs, o)
			addFinding("securitytxt-expires-stale", "security.txt Expires is stale", SeverityLow, ConfidenceHigh, []string{o.ObservationID}, "May reduce trust in security disclosure metadata freshness", "Low", "Update Expires to a future timestamp", "Re-check Expires is in the future", nil)
		}
	}

	out.Observations = obs
	out.CandidateFindings = findings
	return out
}

func parseHost(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(u.Hostname()))
}

func hostInScope(host string, scopeHosts []string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	for _, s := range scopeHosts {
		if h == strings.ToLower(strings.TrimSpace(s)) {
			return true
		}
	}
	return false
}

func extractFieldValue(text, field string) string {
	prefix := strings.ToLower(strings.TrimSpace(field)) + ":"
	for _, ln := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(ln)
		if strings.HasPrefix(strings.ToLower(trim), prefix) {
			parts := strings.SplitN(trim, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func parseSecurityTxtExpires(raw string) (time.Time, bool) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return time.Time{}, false
	}
	formats := []string{time.RFC3339, "2006-01-02"}
	for _, f := range formats {
		if t, err := time.Parse(f, v); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
