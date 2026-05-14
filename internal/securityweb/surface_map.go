package securityweb

import (
	"net/url"
	"sort"
	"strings"
)

func BuildSurfaceMap(results []ExecutionResult, checks []PassiveCheckResult, scopeHosts []string) SurfaceMap {
	m := SurfaceMap{MapID: "surface-map-static-mvp", ScopeHosts: dedupLower(scopeHosts)}

	evidenceSet := map[string]struct{}{}
	hosts := map[string]*SurfaceHost{}
	urls := map[string]*SurfaceURL{}
	paths := map[string]*SurfacePath{}
	signals := map[string]SurfaceSignal{}
	areas := map[string]*SurfaceCandidateArea{}

	addEvidence := func(ev string) {
		ev = strings.TrimSpace(ev)
		if ev != "" {
			evidenceSet[ev] = struct{}{}
		}
	}
	addHost := func(host string, ev string, src SurfaceSource) {
		h := strings.ToLower(strings.TrimSpace(host))
		if h == "" {
			return
		}
		v, ok := hosts[h]
		if !ok {
			v = &SurfaceHost{Host: h, ScopeStatus: ScopeStatusForHost(h, m.ScopeHosts)}
			hosts[h] = v
		}
		v.EvidenceIDs = appendUnique(v.EvidenceIDs, ev)
		v.Sources = appendUniqueSource(v.Sources, src)
	}
	addPath := func(path string, src SurfaceSource, ev string) {
		p := normalizePath(path)
		if p == "" {
			return
		}
		key := string(src) + "|" + p
		v, ok := paths[key]
		if !ok {
			v = &SurfacePath{Path: p, Source: src, SensitivityHint: ClassifyPathSensitivity(p)}
			paths[key] = v
		}
		v.EvidenceIDs = appendUnique(v.EvidenceIDs, ev)
	}
	addURL := func(raw string, ev string, src SurfaceSource) {
		n, h, p, ok := NormalizeSurfaceURL(raw)
		if !ok {
			return
		}
		key := h + p
		v, found := urls[key]
		if !found {
			v = &SurfaceURL{URL: n, Host: h, Path: p, ScopeStatus: ScopeStatusForHost(h, m.ScopeHosts), Source: src}
			urls[key] = v
		}
		v.EvidenceIDs = appendUnique(v.EvidenceIDs, ev)
		addHost(h, ev, src)
		addPath(p, src, ev)
	}

	addSignal := func(s SurfaceSignal) {
		if strings.TrimSpace(s.SignalID) == "" {
			return
		}
		signals[s.SignalID] = s
	}
	addArea := func(id string, cat FindingCategory, reason, next string, conf ConfidenceHint, ev string) {
		if strings.TrimSpace(id) == "" {
			return
		}
		v, ok := areas[id]
		if !ok {
			v = &SurfaceCandidateArea{AreaID: id, Category: cat, Reason: reason, SuggestedNextCheck: next, ConfidenceHint: conf}
			areas[id] = v
		}
		v.EvidenceIDs = appendUnique(v.EvidenceIDs, ev)
	}

	for _, r := range results {
		addEvidence(r.EvidenceID)
		addURL(firstNonEmpty(r.FinalURL, r.URL), r.EvidenceID, SurfaceSourceExecutionResult)
	}

	for _, c := range checks {
		addEvidence(c.EvidenceID)
		for _, o := range c.Observations {
			src := sourceFromCheckID(o.SourceCheckID)
			addURL(o.AffectedURL, o.EvidenceID, src)
			for _, p := range splitObservedValues(o.ObservedValue) {
				if looksLikeURL(p) {
					addURL(p, o.EvidenceID, src)
					continue
				}
				if strings.HasPrefix(strings.TrimSpace(p), "/") {
					addPath(p, src, o.EvidenceID)
				}
			}
			addSignal(SurfaceSignal{SignalID: string(o.SourceCheckID) + ":" + o.ObservationID, Title: o.Title, Category: FindingCategoryInformational, EvidenceID: o.EvidenceID, SourceCheckID: o.SourceCheckID, RelatedObservationIDs: []string{o.ObservationID}, Notes: o.Notes})

			if o.ObservationID == "robots_missing" || o.ObservationID == "sitemap_missing" || o.ObservationID == "securitytxt_missing" {
				addArea("unknown_scope_needs_context", FindingCategoryInformational, "Missing discovery file observed; requires context", "manual-context-review", ConfidenceMedium, o.EvidenceID)
			}
			if o.ObservationID == "securitytxt_contact_present" {
				addArea("security_contact", FindingCategoryInformational, "Security contact channel observed", "validate-security-contact-process", ConfidenceHigh, o.EvidenceID)
			}
			if o.ObservationID == "sitemap_out_of_scope_urls" && len(m.ScopeHosts) > 0 {
				addArea("out_of_scope_reference", FindingCategoryExposure, "Out-of-scope URLs observed in sitemap", "scope-boundary-review", ConfidenceHigh, o.EvidenceID)
			}
			if o.ObservationID == "sitemap_needs_context" && len(m.ScopeHosts) == 0 {
				addArea("unknown_scope_needs_context", FindingCategoryInformational, "Scope hosts missing for sitemap interpretation", "provide-scope-hosts", ConfidenceHigh, o.EvidenceID)
			}
		}
		for _, f := range c.CandidateFindings {
			addSignal(SurfaceSignal{SignalID: string(f.SourceCheckID) + ":" + f.CandidateID, Title: f.Title, Category: f.Category, EvidenceID: f.EvidenceID, SourceCheckID: f.SourceCheckID, RelatedObservationIDs: f.RelatedObservationIDs, Notes: strings.Join(f.Limitations, "; ")})
			src := sourceFromCheckID(f.SourceCheckID)
			addURL(f.AffectedURL, f.EvidenceID, src)
			switch f.SourceCheckID {
			case PassiveCheckIDHeadersMVP:
				addArea("security_headers", FindingCategoryConfiguration, "Header posture signal observed", "passive-headers-review", f.ConfidenceHint, f.EvidenceID)
			case PassiveCheckIDRobotsMVP:
				addArea("exposure_from_robots", FindingCategoryExposure, "Robots-based exposure hints observed", "robots-path-policy-review", f.ConfidenceHint, f.EvidenceID)
			case PassiveCheckIDSitemapMVP:
				addArea("exposure_from_sitemap", FindingCategoryExposure, "Sitemap-based exposure hints observed", "sitemap-exposure-review", f.ConfidenceHint, f.EvidenceID)
			case PassiveCheckIDSecurityTxtMVP:
				addArea("security_contact", FindingCategoryInformational, "security.txt metadata quality signal observed", "securitytxt-governance-review", f.ConfidenceHint, f.EvidenceID)
			}
		}
		m.Limitations = appendUnique(m.Limitations, c.Limitations...)
	}

	if len(m.ScopeHosts) == 0 {
		addArea("unknown_scope_needs_context", FindingCategoryInformational, "Scope hosts were not provided", "provide-scope-hosts", ConfidenceHigh, "")
	}

	m.EvidenceIDs = mapKeys(evidenceSet)
	for _, h := range hosts {
		m.Hosts = append(m.Hosts, *h)
	}
	for _, u := range urls {
		m.URLs = append(m.URLs, *u)
	}
	for _, p := range paths {
		m.Paths = append(m.Paths, *p)
	}
	for _, s := range signals {
		m.Signals = append(m.Signals, s)
	}
	for _, a := range areas {
		m.CandidateAreas = append(m.CandidateAreas, *a)
	}
	sort.Slice(m.Hosts, func(i, j int) bool { return m.Hosts[i].Host < m.Hosts[j].Host })
	sort.Slice(m.URLs, func(i, j int) bool { return m.URLs[i].URL < m.URLs[j].URL })
	sort.Slice(m.Paths, func(i, j int) bool { return m.Paths[i].Path < m.Paths[j].Path })
	sort.Slice(m.Signals, func(i, j int) bool { return m.Signals[i].SignalID < m.Signals[j].SignalID })
	sort.Slice(m.CandidateAreas, func(i, j int) bool { return m.CandidateAreas[i].AreaID < m.CandidateAreas[j].AreaID })
	return m
}


func NormalizeSurfaceURL(raw string) (string, string, string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", "", "", false
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return "", "", "", false
	}
	path := normalizePath(u.EscapedPath())
	if path == "" {
		path = "/"
	}
	n := u.Scheme + "://" + host + path
	return n, host, path, true
}

func ClassifyPathSensitivity(path string) SensitivityHint {
	p := strings.ToLower(path)
	switch {
	case strings.Contains(p, "admin"):
		return SensitivityAdmin
	case strings.Contains(p, "backup"):
		return SensitivityBackup
	case strings.Contains(p, "private"):
		return SensitivityPrivate
	case strings.Contains(p, "staging"):
		return SensitivityStaging
	case strings.Contains(p, "auth") || strings.Contains(p, "login"):
		return SensitivityAuth
	default:
		if strings.TrimSpace(p) == "" {
			return SensitivityUnknown
		}
		return SensitivityNone
	}
}

func ScopeStatusForHost(host string, scopeHosts []string) SurfaceScopeStatus {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return SurfaceScopeUnknownScope
	}
	if len(scopeHosts) == 0 {
		return SurfaceScopeUnknownScope
	}
	for _, s := range scopeHosts {
		if h == strings.ToLower(strings.TrimSpace(s)) {
			return SurfaceScopeInScope
		}
	}
	return SurfaceScopeOutOfScope
}

func normalizePath(p string) string {
	pp := strings.TrimSpace(p)
	if pp == "" {
		return "/"
	}
	if !strings.HasPrefix(pp, "/") {
		pp = "/" + pp
	}
	return pp
}

func sourceFromCheckID(id PassiveCheckID) SurfaceSource {
	switch id {
	case PassiveCheckIDHeadersMVP:
		return SurfaceSourceHeaders
	case PassiveCheckIDRobotsMVP:
		return SurfaceSourceRobots
	case PassiveCheckIDSitemapMVP:
		return SurfaceSourceSitemap
	case PassiveCheckIDSecurityTxtMVP:
		return SurfaceSourceSecurityTxt
	default:
		return SurfaceSourceExecutionResult
	}
}

func splitObservedValues(v string) []string {
	out := []string{}
	for _, p := range strings.Split(v, ",") {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func looksLikeURL(v string) bool {
	vv := strings.ToLower(strings.TrimSpace(v))
	return strings.HasPrefix(vv, "http://") || strings.HasPrefix(vv, "https://")
}

func dedupLower(in []string) []string {
	set := map[string]struct{}{}
	for _, v := range in {
		t := strings.ToLower(strings.TrimSpace(v))
		if t != "" {
			set[t] = struct{}{}
		}
	}
	return mapKeys(set)
}

func mapKeys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func appendUnique(in []string, values ...string) []string {
	set := map[string]struct{}{}
	for _, v := range in {
		set[v] = struct{}{}
	}
	for _, v := range values {
		vv := strings.TrimSpace(v)
		if vv == "" {
			continue
		}
		set[vv] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func appendUniqueSource(in []SurfaceSource, value SurfaceSource) []SurfaceSource {
	for _, v := range in {
		if v == value {
			return in
		}
	}
	return append(in, value)
}
