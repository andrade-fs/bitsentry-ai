package securitywebhttp_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bitsentry-ai/internal/securityweb"
	"bitsentry-ai/internal/securitywebhttp"
)

func TestE2EPassivePipeline_HappyCoverage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/":
			w.Header().Set("X-Content-Type-Options", "nosniff") // missing CSP intentionally
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/robots.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin\nSitemap: https://example.com/sitemap.xml\n"))
		case r.Method == http.MethodGet && r.URL.Path == "/sitemap.xml":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<urlset><url><loc>https://example.com/admin</loc></url></urlset>"))
		case r.Method == http.MethodGet && r.URL.Path == "/.well-known/security.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Contact: mailto:security@example.com\nExpires: 2099-01-01T00:00:00Z\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	plan := securityweb.WebTestPlan{
		PlanID: "web-test-plan-static-mvp",
		TestPlanItems: []securityweb.WebTestPlanItem{
			{ItemID: "e2e-h", HypothesisID: "hyp-h", SuggestedNextCheck: securityweb.NextHeadersHardeningReview, Status: securityweb.WebTestPlanStatusPlanned},
			{ItemID: "e2e-r", HypothesisID: "hyp-r", SuggestedNextCheck: securityweb.NextRobotsExposureReview, Status: securityweb.WebTestPlanStatusPlanned},
			{ItemID: "e2e-s", HypothesisID: "hyp-s", SuggestedNextCheck: securityweb.NextSitemapExposureReview, Status: securityweb.WebTestPlanStatusPlanned},
			{ItemID: "e2e-t", HypothesisID: "hyp-t", SuggestedNextCheck: securityweb.NextSecurityTxtGovernance, Status: securityweb.WebTestPlanStatusPlanned},
		},
	}

	bridge := securityweb.BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	if bridge.ExecutionMode != securityweb.ExecutionModeDryRun || bridge.WouldExecute {
		t.Fatalf("bridge plan must stay dry_run and would_execute=false")
	}

	tr, err := securitywebhttp.New(400*time.Millisecond, 4096)
	if err != nil {
		t.Fatalf("transport new: %v", err)
	}
	exec := securityweb.NewOfflineControlledExecutor(tr, securityweb.DefaultRedactor{})
	execCtxLocal := execCtx(ts.URL)
	execCtxLocal.RequestBudget = 10

	var results []securityweb.ExecutionResult
	var checks []securityweb.PassiveCheckResult
	for _, req := range bridge.PlannedRequests {
		req = remapRequestToHTTptest(t, req, ts.URL)
		approval := buildApprovalForRequest(req, execCtxLocal)
		approval.MaxRequests = execCtxLocal.RequestBudget
		res := exec.ExecuteApproved(execCtxLocal, req, approval)
		if res.PolicyDecision != "allow" {
			t.Fatalf("expected allow for %s %s, got policy=%s violations=%v", req.Method, req.URL, res.PolicyDecision, res.Violations)
		}
		if res.StatusCode < 200 || res.StatusCode >= 400 {
			t.Fatalf("expected 2xx/3xx for %s %s, got status=%d", req.Method, req.URL, res.StatusCode)
		}
		results = append(results, res)
		if req.Method == securityweb.MethodHEAD && strings.HasSuffix(req.URL, "/") {
			checks = append(checks, securityweb.EvaluatePassiveHeaders(securityweb.HeaderCheckInput{ExecutionResult: res, RequestedMethod: securityweb.MethodHEAD}))
		} else if strings.Contains(req.URL, "/robots.txt") {
			checks = append(checks, securityweb.EvaluatePassiveRobots(res))
		} else if strings.Contains(req.URL, "/sitemap.xml") {
			checks = append(checks, securityweb.EvaluatePassiveSitemap(res, []string{"example.com"}))
		} else if strings.Contains(req.URL, "/.well-known/security.txt") {
			checks = append(checks, securityweb.EvaluatePassiveSecurityTxt(res, func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }))
		}
	}

	if len(checks) != 4 {
		t.Fatalf("expected 4 passive check results, got %d", len(checks))
	}

	surface := securityweb.BuildSurfaceMap(results, checks, []string{"example.com"})
	if len(surface.Hosts) == 0 || len(surface.URLs) == 0 || len(surface.Paths) == 0 || len(surface.Signals) == 0 {
		t.Fatalf("expected populated surface map")
	}
	for _, area := range []string{"security_headers", "exposure_from_robots", "exposure_from_sitemap", "security_contact"} {
		if !hasArea(surface, area) {
			t.Fatalf("expected candidate area %s, got=%v", area, areaIDs(surface))
		}
	}

	risk := securityweb.BuildRiskHypothesesFromSurfaceMap(surface)
	if len(risk.Hypotheses) == 0 {
		t.Fatalf("expected risk hypotheses")
	}
	if !riskHasSkill(risk, securityweb.SkillHeadersSecurityReview) ||
		!riskHasSkill(risk, securityweb.SkillExposureReview) ||
		!riskHasSkill(risk, securityweb.SkillAccessControlReview) ||
		!riskHasSkill(risk, securityweb.SkillSecurityTxtReview) {
		t.Fatalf("expected required expert skills in hypotheses")
	}

	bridgeCtxFinal := mkBridgeCtx()
	finalPlan := securityweb.BuildWebTestPlanFromHypotheses(risk, &bridgeCtxFinal)
	if finalPlan.ExecutionMode != securityweb.ExecutionModePlanningOnly {
		t.Fatalf("final web test plan must remain planning_only")
	}
	if len(finalPlan.TestPlanItems) == 0 {
		t.Fatalf("expected final test plan items")
	}
	for _, h := range risk.Hypotheses {
		for _, n := range h.SuggestedNextChecks {
			if !strings.Contains(string(n), "dry-run") {
				t.Fatalf("hypothesis next check must remain dry-run/planning-only: %s", n)
			}
		}
	}
}

func TestE2EPassivePipeline_TraceabilityContracts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/robots.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin\n"))
		case r.Method == http.MethodGet && r.URL.Path == "/sitemap.xml":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<urlset><url><loc>https://example.com/admin</loc></url></urlset>"))
		case r.Method == http.MethodGet && r.URL.Path == "/.well-known/security.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Contact: mailto:security@example.com\nExpires: 2099-01-01T00:00:00Z\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	plan := securityweb.WebTestPlan{
		PlanID: "web-test-plan-static-mvp",
		TestPlanItems: []securityweb.WebTestPlanItem{
			{ItemID: "trace-h", HypothesisID: "hyp-h", SuggestedNextCheck: securityweb.NextHeadersHardeningReview, Status: securityweb.WebTestPlanStatusPlanned},
			{ItemID: "trace-r", HypothesisID: "hyp-r", SuggestedNextCheck: securityweb.NextRobotsExposureReview, Status: securityweb.WebTestPlanStatusPlanned},
			{ItemID: "trace-s", HypothesisID: "hyp-s", SuggestedNextCheck: securityweb.NextSitemapExposureReview, Status: securityweb.WebTestPlanStatusPlanned},
			{ItemID: "trace-t", HypothesisID: "hyp-t", SuggestedNextCheck: securityweb.NextSecurityTxtGovernance, Status: securityweb.WebTestPlanStatusPlanned},
		},
	}
	bridge := securityweb.BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	if bridge.ExecutionMode != securityweb.ExecutionModeDryRun || bridge.WouldExecute {
		t.Fatalf("bridge must remain non-operational")
	}

	tr, _ := securitywebhttp.New(400*time.Millisecond, 4096)
	exec := securityweb.NewOfflineControlledExecutor(tr, securityweb.DefaultRedactor{})
	xctx := execCtx(ts.URL)
	xctx.RequestBudget = 10

	results := []securityweb.ExecutionResult{}
	checks := []securityweb.PassiveCheckResult{}
	for _, req := range bridge.PlannedRequests {
		req = remapRequestToHTTptest(t, req, ts.URL)
		approval := buildApprovalForRequest(req, xctx)
		approval.MaxRequests = xctx.RequestBudget
		res := exec.ExecuteApproved(xctx, req, approval)
		if res.PolicyDecision != "allow" {
			t.Fatalf("expected allow for %s %s, got policy=%s violations=%v", req.Method, req.URL, res.PolicyDecision, res.Violations)
		}
		if res.RequestID != req.RequestRef || res.ApprovalID != approval.ApprovalID || res.EvidenceID == "" {
			t.Fatalf("traceability broken request->approval->execution")
		}
		results = append(results, res)

		var chk securityweb.PassiveCheckResult
		switch {
		case req.Method == securityweb.MethodHEAD:
			chk = securityweb.EvaluatePassiveHeaders(securityweb.HeaderCheckInput{ExecutionResult: res, RequestedMethod: securityweb.MethodHEAD})
		case strings.Contains(req.URL, "/robots.txt"):
			chk = securityweb.EvaluatePassiveRobots(res)
		case strings.Contains(req.URL, "/sitemap.xml"):
			chk = securityweb.EvaluatePassiveSitemap(res, []string{"example.com"})
		case strings.Contains(req.URL, "/.well-known/security.txt"):
			chk = securityweb.EvaluatePassiveSecurityTxt(res, nil)
		}
		if chk.EvidenceID != res.EvidenceID {
			t.Fatalf("passive check evidence must match execution evidence")
		}
		for _, o := range chk.Observations {
			if o.SourceCheckID == "" {
				t.Fatalf("observation must preserve source check id")
			}
		}
		for _, f := range chk.CandidateFindings {
			if len(f.RelatedObservationIDs) == 0 {
				t.Fatalf("candidate findings must preserve related observation ids")
			}
		}
		checks = append(checks, chk)
	}

	surface := securityweb.BuildSurfaceMap(results, checks, []string{"example.com"})
	if len(surface.EvidenceIDs) == 0 || len(surface.Signals) == 0 || len(surface.CandidateAreas) == 0 {
		t.Fatalf("surface traceability artifacts missing")
	}

	risk := securityweb.BuildRiskHypothesesFromSurfaceMap(surface)
	if len(risk.EvidenceIDs) == 0 || len(risk.Hypotheses) == 0 {
		t.Fatalf("risk hypotheses traceability missing")
	}
	for _, h := range risk.Hypotheses {
		if len(h.SourceCandidateAreaIDs) == 0 {
			t.Fatalf("hypothesis must preserve source candidate area ids")
		}
	}

	bridgeCtxFinal := mkBridgeCtx()
	finalPlan := securityweb.BuildWebTestPlanFromHypotheses(risk, &bridgeCtxFinal)
	if finalPlan.ExecutionMode != securityweb.ExecutionModePlanningOnly {
		t.Fatalf("final plan must remain planning_only")
	}
	for _, it := range finalPlan.TestPlanItems {
		if it.HypothesisID == "" || len(it.EvidenceIDs) == 0 {
			t.Fatalf("test plan item must preserve hypothesis/evidence linkage")
		}
		if !strings.Contains(string(it.SuggestedNextCheck), "dry-run") {
			t.Fatalf("final suggested next checks must remain dry-run/planning-only")
		}
	}
}

func hasArea(m securityweb.SurfaceMap, id string) bool {
	for _, a := range m.CandidateAreas {
		if a.AreaID == id {
			return true
		}
	}
	return false
}

func riskHasSkill(set securityweb.RiskHypothesisSet, s securityweb.SuggestedExpertSkill) bool {
	for _, h := range set.Hypotheses {
		for _, ss := range h.SuggestedExpertSkills {
			if ss == s {
				return true
			}
		}
	}
	return false
}


func areaIDs(m securityweb.SurfaceMap) []string {
	out := make([]string, 0, len(m.CandidateAreas))
	for _, a := range m.CandidateAreas {
		out = append(out, a.AreaID)
	}
	return out
}
