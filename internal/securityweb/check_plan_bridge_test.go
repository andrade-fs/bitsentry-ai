package securityweb

import "testing"

func TestBridge_HeadersToHeadRoot(t *testing.T) {
	plan := mkPlanItem("plan-item-1", "h1", NextHeadersHardeningReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := mustRequestByItem(t, out, "plan-item-1")
	if req.Method != MethodHEAD || req.URL != "https://example.com/" {
		t.Fatalf("expected HEAD / mapping")
	}
}

func TestBridge_RobotsToGetRobots(t *testing.T) {
	plan := mkPlanItem("plan-item-2", "h2", NextRobotsExposureReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := mustRequestByItem(t, out, "plan-item-2")
	if req.Method != MethodGET || req.URL != "https://example.com/robots.txt" {
		t.Fatalf("expected GET /robots.txt mapping")
	}
}

func TestBridge_SitemapToGetSitemap(t *testing.T) {
	plan := mkPlanItem("plan-item-3", "h3", NextSitemapExposureReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := mustRequestByItem(t, out, "plan-item-3")
	if req.Method != MethodGET || req.URL != "https://example.com/sitemap.xml" {
		t.Fatalf("expected GET /sitemap.xml mapping")
	}
}

func TestBridge_SecurityTxtToGetWellKnown(t *testing.T) {
	plan := mkPlanItem("plan-item-4", "h4", NextSecurityTxtGovernance, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := mustRequestByItem(t, out, "plan-item-4")
	if req.Method != MethodGET || req.URL != "https://example.com/.well-known/security.txt" {
		t.Fatalf("expected GET /.well-known/security.txt mapping")
	}
}

func TestBridge_AuthAccessItemsNotConverted(t *testing.T) {
	plan := WebTestPlan{PlanID: "web-test-plan-static-mvp", TestPlanItems: []WebTestPlanItem{
		{ItemID: "plan-item-a", HypothesisID: "ha", SuggestedNextCheck: NextAuthBoundaryReview, Status: WebTestPlanStatusPlanned},
		{ItemID: "plan-item-b", HypothesisID: "hb", SuggestedNextCheck: NextAccessControlReviewDryRun, Status: WebTestPlanStatusPlanned},
	}}
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	if len(out.PlannedRequests) != 0 {
		t.Fatalf("expected no planned requests for non-convertible items")
	}
	assertViolationCode(t, out, "bridge_item_not_convertible_yet")
}

func TestBridge_MissingBaseTargetBlocksBridge(t *testing.T) {
	plan := mkPlanItem("plan-item-1", "h1", NextHeadersHardeningReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "")
	assertViolationCode(t, out, "bridge_base_target_missing")
}

func TestBridge_MissingScopeProducesViolationOrLimitation(t *testing.T) {
	plan := mkPlanItem("plan-item-1", "h1", NextHeadersHardeningReview, WebTestPlanStatusPlanned)
	ctx := mkBridgeCtx()
	ctx.InScopeTargets = nil
	out := BuildControlledCheckPlanFromWebTestPlan(plan, ctx, "example.com")
	assertViolationCode(t, out, "bridge_scope_missing")
}

func TestBridge_AllPlannedRequestsAreGetOrHeadAndWouldExecuteFalse(t *testing.T) {
	plan := WebTestPlan{PlanID: "web-test-plan-static-mvp", TestPlanItems: []WebTestPlanItem{
		{ItemID: "i1", HypothesisID: "h1", SuggestedNextCheck: NextHeadersHardeningReview, Status: WebTestPlanStatusPlanned},
		{ItemID: "i2", HypothesisID: "h2", SuggestedNextCheck: NextRobotsExposureReview, Status: WebTestPlanStatusPlanned},
	}}
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	if out.WouldExecute {
		t.Fatalf("bridge must be non-executing")
	}
	for _, r := range out.PlannedRequests {
		if r.Method != MethodGET && r.Method != MethodHEAD {
			t.Fatalf("only GET/HEAD allowed")
		}
	}
}

func TestBridge_StableRequestRefAndLinkToItem(t *testing.T) {
	plan := mkPlanItem("plan-item-z", "hz", NextHeadersHardeningReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := mustRequestByItem(t, out, "plan-item-z")
	if req.RequestRef != "bridge-plan-item-z" {
		t.Fatalf("expected stable request ref")
	}
}

func TestBridge_RequiredApprovalEvidenceAndPolicyPerRequest(t *testing.T) {
	plan := mkPlanItem("plan-item-1", "h1", NextHeadersHardeningReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	if len(out.RequiredApprovals) != 1 {
		t.Fatalf("expected one future-facing required approval")
	}
	if len(out.EvidenceTemplates) != 1 {
		t.Fatalf("expected one evidence template")
	}
	if len(out.PolicyDecisions) != 1 {
		t.Fatalf("expected one policy decision")
	}
}

func TestBridge_DryRunPolicyDenyDoesNotRemovePlannedArtifact(t *testing.T) {
	plan := mkPlanItem("plan-item-1", "h1", NextHeadersHardeningReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	if len(out.PlannedRequests) != 1 {
		t.Fatalf("expected planned request even with dry_run deny")
	}
	if out.PolicyDecisions[0].Decision.Allowed {
		t.Fatalf("dry_run policy should deny execution")
	}
}

func TestBridge_NoPostNoQueryPayload(t *testing.T) {
	plan := mkPlanItem("plan-item-1", "h1", NextSitemapExposureReview, WebTestPlanStatusPlanned)
	out := BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := mustRequestByItem(t, out, "plan-item-1")
	if req.Method == MethodPOST {
		t.Fatalf("no POST allowed")
	}
	if req.Headers == nil {
		return
	}
	if _, ok := req.Headers["Content-Type"]; ok {
		t.Fatalf("no payload headers expected")
	}
}

func mkPlanItem(itemID, hypID string, next SuggestedNextCheck, status WebTestPlanStatus) WebTestPlan {
	return WebTestPlan{PlanID: "web-test-plan-static-mvp", TestPlanItems: []WebTestPlanItem{{ItemID: itemID, HypothesisID: hypID, SuggestedNextCheck: next, Status: status}}}
}

func mkBridgeCtx() AssessmentSessionContext {
	return AssessmentSessionContext{
		SessionID:            "sess-1",
		AuthorizationRef:     "auth-1",
		ScopeRef:             "scope-1",
		InScopeTargets:       []string{"example.com"},
		ExecutionMode:        ExecutionModePlanningOnly,
		Intensity:            IntensityLow,
		RateLimitPerMinute:   10,
		RequestBudget:        5,
		TimeoutSeconds:       10,
		MaxResponseSizeBytes: 1024,
		MaxPreviewSizeBytes:  256,
		StopConditions:       []string{"user_stop"},
		EvidencePlanRef:      "ev-plan-1",
	}
}

func mustRequestByItem(t *testing.T, out ControlledCheckPlan, itemID string) PlannedRequest {
	t.Helper()
	ref := "bridge-" + itemID
	for _, r := range out.PlannedRequests {
		if r.RequestRef == ref {
			return r
		}
	}
	t.Fatalf("request for item %s not found", itemID)
	return PlannedRequest{}
}

func assertViolationCode(t *testing.T, out ControlledCheckPlan, code string) {
	t.Helper()
	for _, v := range out.Violations {
		if v.Code == code {
			return
		}
	}
	t.Fatalf("expected violation code %s", code)
}
