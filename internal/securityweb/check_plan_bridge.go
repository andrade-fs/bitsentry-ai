package securityweb

import (
	"fmt"
	"sort"
	"strings"
)

func BuildControlledCheckPlanFromWebTestPlan(plan WebTestPlan, ctx AssessmentSessionContext, baseTarget string) ControlledCheckPlan {
	out := ControlledCheckPlan{
		PlanID:            "controlled-check-plan-static-mvp",
		FromWebTestPlanID: plan.PlanID,
		ExecutionMode:     ExecutionModeDryRun,
		WouldExecute:      false,
		Limitations: []string{
			"Dry-run planning artifact only; no execution performed",
			"safe_probe classification does not imply transport/executor invocation",
		},
	}

	base := strings.TrimSpace(baseTarget)
	if base == "" {
		out.Violations = append(out.Violations, RequestPlanBridgeViolation{ItemID: "plan", Code: "bridge_base_target_missing", Reason: "base target is required", Field: "base_target"})
		out.Limitations = appendUnique(out.Limitations, "Bridge blocked: missing base target")
	}
	if len(ctx.InScopeTargets) == 0 {
		out.Violations = append(out.Violations, RequestPlanBridgeViolation{ItemID: "plan", Code: "bridge_scope_missing", Reason: "in-scope targets are required for controlled planning", Field: "in_scope_targets"})
		out.Limitations = appendUnique(out.Limitations, "Scope is missing or unknown; some items remain blocked")
	}

	for _, item := range plan.TestPlanItems {
		cItem := ControlledCheckPlanItem{TestPlanItemID: item.ItemID}
		if base == "" {
			cItem.Planned = false
			cItem.Reason = "base target missing"
			out.PlannedItems = append(out.PlannedItems, cItem)
			continue
		}

		method, path, convertible := mapBridgeMethodPath(item.SuggestedNextCheck)
		if !convertible {
			cItem.Planned = false
			cItem.Reason = "bridge_item_not_convertible_yet"
			out.Violations = append(out.Violations, RequestPlanBridgeViolation{ItemID: item.ItemID, Code: "bridge_item_not_convertible_yet", Reason: "suggested next check not supported in MVP bridge", Field: "suggested_next_check"})
			out.PlannedItems = append(out.PlannedItems, cItem)
			continue
		}

		if item.Status == WebTestPlanStatusBlockedNeedsScope {
			cItem.Planned = false
			cItem.Reason = "blocked_needs_scope"
			out.Violations = append(out.Violations, RequestPlanBridgeViolation{ItemID: item.ItemID, Code: "bridge_item_blocked_needs_scope", Reason: "item is blocked pending scope/context", Field: "status"})
			out.PlannedItems = append(out.PlannedItems, cItem)
			continue
		}

		reqRef := "bridge-" + item.ItemID
		req := PlannedRequest{
			RequestRef: reqRef,
			Target:     base,
			URL:        fmt.Sprintf("https://%s%s", base, path),
			Method:     method,
			ToolClass:  ToolClassSafeProbe,
			Headers:    map[string]string{},
		}
		if strings.Contains(req.URL, "?") {
			out.Violations = append(out.Violations, RequestPlanBridgeViolation{ItemID: item.ItemID, Code: "bridge_query_not_allowed", Reason: "query params are not allowed in MVP bridge", Field: "url"})
			cItem.Planned = false
			cItem.Reason = "query_not_allowed"
			out.PlannedItems = append(out.PlannedItems, cItem)
			continue
		}
		if method != MethodGET && method != MethodHEAD {
			out.Violations = append(out.Violations, RequestPlanBridgeViolation{ItemID: item.ItemID, Code: "bridge_method_not_allowed", Reason: "only GET/HEAD allowed", Field: "method"})
			cItem.Planned = false
			cItem.Reason = "method_not_allowed"
			out.PlannedItems = append(out.PlannedItems, cItem)
			continue
		}

		policyCtx := ctx
		policyCtx.ExecutionMode = ExecutionModeDryRun
		if len(policyCtx.StopConditions) == 0 {
			policyCtx.StopConditions = append([]string{}, plan.StopConditions...)
		}
		if strings.TrimSpace(policyCtx.EvidencePlanRef) == "" {
			policyCtx.EvidencePlanRef = "bridge-evidence-plan-static-mvp"
		}
		if !containsString(policyCtx.InScopeTargets, base) {
			policyCtx.InScopeTargets = append(policyCtx.InScopeTargets, base)
		}
		pd, violations := DefaultPolicyEvaluator{}.Evaluate(policyCtx, req)
		out.PolicyDecisions = append(out.PolicyDecisions, RequestPlanPolicyDecision{ItemID: item.ItemID, RequestRef: reqRef, Decision: pd})
		for _, v := range violations {
			out.Violations = append(out.Violations, RequestPlanBridgeViolation{ItemID: item.ItemID, Code: v.Code, Reason: v.Reason, Field: v.Field})
		}

		out.PlannedRequests = append(out.PlannedRequests, req)
		out.EvidenceTemplates = append(out.EvidenceTemplates, EvidenceEntry{
			EvidenceID:        "WEB-EV-BRIDGE-" + reqRef,
			SessionMode:       fmt.Sprintf("%s / %s", policyCtx.SessionID, policyCtx.ExecutionMode),
			AuthorizationRef:  policyCtx.AuthorizationRef,
			ScopeRef:          policyCtx.ScopeRef,
			PlannedRequestRef: reqRef,
			PolicyDecision:    fmt.Sprintf("%t (%s)", pd.Allowed, pd.Reason),
			RedactionApplied:  false,
			NotesAssumptions:  "dry-run planning template only",
			RequestURL:        req.URL,
			RequestHeaders:    req.Headers,
		})
		out.RequiredApprovals = append(out.RequiredApprovals, RequiredApproval{
			ApprovalID:    "future-approval-" + reqRef,
			HypothesisID:  item.HypothesisID,
			Reason:        "Future execute_approved requires explicit approval",
			RequiredMode:  ExecutionModeExecuteApproved,
			RequiredScope: "exact_scope_required",
			Notes: []string{
				"Future-facing only; no execution in this phase",
				"test_plan_item_id=" + item.ItemID,
				"request_ref=" + reqRef,
				"method=" + string(method),
				"url=" + req.URL,
			},
		})
		cItem.Planned = true
		cItem.RequestRef = reqRef
		cItem.Reason = "planned_dry_run_artifact"
		out.PlannedItems = append(out.PlannedItems, cItem)
	}

	sort.Slice(out.PlannedItems, func(i, j int) bool { return out.PlannedItems[i].TestPlanItemID < out.PlannedItems[j].TestPlanItemID })
	sort.Slice(out.PlannedRequests, func(i, j int) bool { return out.PlannedRequests[i].RequestRef < out.PlannedRequests[j].RequestRef })
	sort.Slice(out.RequiredApprovals, func(i, j int) bool { return out.RequiredApprovals[i].ApprovalID < out.RequiredApprovals[j].ApprovalID })
	sort.Slice(out.PolicyDecisions, func(i, j int) bool { return out.PolicyDecisions[i].RequestRef < out.PolicyDecisions[j].RequestRef })
	sort.Slice(out.EvidenceTemplates, func(i, j int) bool { return out.EvidenceTemplates[i].PlannedRequestRef < out.EvidenceTemplates[j].PlannedRequestRef })
	return out
}

func mapBridgeMethodPath(next SuggestedNextCheck) (RequestMethod, string, bool) {
	switch next {
	case NextHeadersHardeningReview:
		return MethodHEAD, "/", true
	case NextRobotsExposureReview:
		return MethodGET, "/robots.txt", true
	case NextSitemapExposureReview:
		return MethodGET, "/sitemap.xml", true
	case NextSecurityTxtGovernance:
		return MethodGET, "/.well-known/security.txt", true
	default:
		return "", "", false
	}
}

func containsString(arr []string, v string) bool {
	for _, x := range arr {
		if strings.EqualFold(strings.TrimSpace(x), strings.TrimSpace(v)) {
			return true
		}
	}
	return false
}
