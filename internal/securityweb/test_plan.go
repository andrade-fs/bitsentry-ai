package securityweb

import (
	"sort"
	"strings"
)

func BuildWebTestPlanFromHypotheses(h RiskHypothesisSet, ctx *AssessmentSessionContext) WebTestPlan {
	execMode, intensity, budget, rateLimit, stop := DefaultWebTestPlanLimits(ctx)
	plan := WebTestPlan{
		PlanID:              "web-test-plan-static-mvp",
		FromHypothesisSetID: h.SetID,
		ExecutionMode:       execMode,
		Intensity:           intensity,
		RequestBudget:       budget,
		RateLimitPerMinute:  rateLimit,
		StopConditions:      stop,
		Limitations: []string{
			"Planning-only artifact; no execution performed",
			"Priority is triage order and not vulnerability severity",
			"ProposedMethod intentionally empty in MVP to avoid request-ready interpretation",
		},
	}

	for _, rh := range h.Hypotheses {
		item := buildPlanItem(rh, execMode, stop)
		plan.TestPlanItems = append(plan.TestPlanItems, item)
		if item.Status == WebTestPlanStatusBlockedNeedsApproval {
			plan.RequiredApprovals = append(plan.RequiredApprovals, RequiredApproval{
				ApprovalID:    "approval-needed-" + item.ItemID,
				HypothesisID:  item.HypothesisID,
				Reason:        "Future execution would require explicit approval",
				RequiredMode:  ExecutionModeExecuteApproved,
				RequiredScope: "exact_scope_required",
				Notes:         []string{"Future-facing only; this plan does not execute requests"},
			})
		}
	}

	sort.Slice(plan.TestPlanItems, func(i, j int) bool { return plan.TestPlanItems[i].ItemID < plan.TestPlanItems[j].ItemID })
	sort.Slice(plan.RequiredApprovals, func(i, j int) bool { return plan.RequiredApprovals[i].ApprovalID < plan.RequiredApprovals[j].ApprovalID })
	plan.Limitations = appendUnique(plan.Limitations, h.Limitations...)
	return plan
}

func DefaultWebTestPlanLimits(ctx *AssessmentSessionContext) (ExecutionMode, Intensity, int, int, []string) {
	mode := ExecutionModePlanningOnly
	intensity := IntensityLow
	budget := 0
	rate := 0
	if ctx != nil {
		if ctx.RateLimitPerMinute > 0 {
			rate = ctx.RateLimitPerMinute
		}
		if len(ctx.InScopeTargets) > 0 {
			intensity = IntensityLow
		}
	}
	stop := []string{"unexpected_5xx", "rate_limit_hit", "out_of_scope_detected", "user_stop", "sensitive_data_detected"}
	return mode, intensity, budget, rate, stop
}

func buildPlanItem(rh RiskHypothesis, mode ExecutionMode, stop []string) WebTestPlanItem {
	item := WebTestPlanItem{
		ItemID:                "plan-item-" + sanitizeID(rh.HypothesisID),
		HypothesisID:          rh.HypothesisID,
		Title:                 rh.Title,
		Category:              rh.Category,
		Priority:              rh.Priority,
		SuggestedExpertSkills: rh.SuggestedExpertSkills,
		ExecutionMode:         mode,
		ToolClass:             "manual",
		ProposedMethod:        "",
		ProposedTarget:        firstNonEmpty(firstFrom(rh.AffectedURLs), firstFrom(rh.AffectedPaths)),
		RequiresApproval:      false,
		ExpectedEvidence:      expectedEvidenceForHypothesis(rh),
		SafetyNotes: []string{
			"planning_only artifact",
			"no active checks",
			"no HTTP requests",
		},
		StopConditions: append([]string{}, stop...),
		EvidenceIDs:    appendUnique(rh.EvidenceIDs),
		Status:         WebTestPlanStatusPlanned,
	}

	if len(rh.SuggestedNextChecks) > 0 {
		item.SuggestedNextCheck = rh.SuggestedNextChecks[0]
	}

	if rh.Status == RiskHypothesisNeedsMoreEvidence {
		item.Status = WebTestPlanStatusBlockedNeedsScope
		item.SafetyNotes = appendUnique(item.SafetyNotes, "blocked pending scope/context clarification")
	}

	if rh.HypothesisID == "rh-out-of-scope-reference" {
		item.Status = WebTestPlanStatusBlockedNeedsScope
		item.SuggestedNextCheck = NextScopeClarification
		item.ProposedTarget = ""
		item.SafetyNotes = appendUnique(item.SafetyNotes, "do not interact with out-of-scope assets")
	}

	if rh.HypothesisID == "rh-unknown-scope-context" {
		item.Status = WebTestPlanStatusBlockedNeedsScope
		item.SuggestedNextCheck = NextScopeClarification
	}

	if requiresAccessControlPlanning(rh) {
		item.SafetyNotes = appendUnique(item.SafetyNotes, "auth/access-control review is planning-only")
	}

	if item.SuggestedNextCheck == "" {
		item.SuggestedNextCheck = NextScopeClarification
	}

	if !strings.Contains(string(item.SuggestedNextCheck), "dry-run") {
		item.SafetyNotes = appendUnique(item.SafetyNotes, "next check treated as planning token")
	}

	return item
}

func expectedEvidenceForHypothesis(rh RiskHypothesis) string {
	if len(rh.EvidenceIDs) == 0 {
		return "Consolidated planning note with hypothesis trace links"
	}
	return "Planning evidence linked to passive hypothesis IDs and existing evidence references"
}

func requiresAccessControlPlanning(rh RiskHypothesis) bool {
	for _, s := range rh.SuggestedExpertSkills {
		if s == SkillAuthSecurityReview || s == SkillAccessControlReview {
			return true
		}
	}
	return false
}

func sanitizeID(v string) string {
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", ":", "-")
	out := replacer.Replace(strings.TrimSpace(v))
	if out == "" {
		return "unknown"
	}
	return out
}

func firstFrom(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[0]
}
