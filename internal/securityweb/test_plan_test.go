package securityweb

import "testing"

func TestWebTestPlan_DeterministicPlanIDAndMode(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp"}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	if p.PlanID != "web-test-plan-static-mvp" {
		t.Fatalf("expected deterministic plan id")
	}
	if p.ExecutionMode != ExecutionModePlanningOnly {
		t.Fatalf("expected global planning_only mode")
	}
}

func TestWebTestPlan_SecurityHeadersItem(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-security-headers-gap", Title: "Security headers hardening gap", Category: FindingCategoryConfiguration, Priority: RiskPriorityMedium, SuggestedExpertSkills: []SuggestedExpertSkill{SkillHeadersSecurityReview}, SuggestedNextChecks: []SuggestedNextCheck{NextHeadersHardeningReview}, EvidenceIDs: []string{"ev-1"}}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	it := planItemByHypothesis(p, "rh-security-headers-gap")
	if it == nil || it.SuggestedNextCheck != NextHeadersHardeningReview {
		t.Fatalf("expected passive headers planning item")
	}
}

func TestWebTestPlan_RobotsAndSitemapExposureItems(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{
		{HypothesisID: "rh-exposure-robots", Title: "robots", SuggestedExpertSkills: []SuggestedExpertSkill{SkillExposureReview, SkillAccessControlReview}, SuggestedNextChecks: []SuggestedNextCheck{NextRobotsExposureReview}, Priority: RiskPriorityLow},
		{HypothesisID: "rh-exposure-sitemap", Title: "sitemap", SuggestedExpertSkills: []SuggestedExpertSkill{SkillExposureReview}, SuggestedNextChecks: []SuggestedNextCheck{NextSitemapExposureReview}, Priority: RiskPriorityLow},
	}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	if planItemByHypothesis(p, "rh-exposure-robots") == nil || planItemByHypothesis(p, "rh-exposure-sitemap") == nil {
		t.Fatalf("expected robots and sitemap exposure items")
	}
}

func TestWebTestPlan_SecurityContactItem(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-security-contact-posture", Title: "contact", SuggestedExpertSkills: []SuggestedExpertSkill{SkillSecurityTxtReview}, SuggestedNextChecks: []SuggestedNextCheck{NextSecurityTxtGovernance}}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	it := planItemByHypothesis(p, "rh-security-contact-posture")
	if it == nil || it.SuggestedNextCheck != NextSecurityTxtGovernance {
		t.Fatalf("expected securitytxt governance item")
	}
}

func TestWebTestPlan_UnknownScopeBlockedNeedsScope(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-unknown-scope-context", Status: RiskHypothesisNeedsMoreEvidence, SuggestedNextChecks: []SuggestedNextCheck{NextScopeClarification}}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	it := planItemByHypothesis(p, "rh-unknown-scope-context")
	if it == nil || it.Status != WebTestPlanStatusBlockedNeedsScope {
		t.Fatalf("expected blocked_needs_scope")
	}
}

func TestWebTestPlan_OutOfScopeBlockedScopeClarificationNoTarget(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-out-of-scope-reference", Status: RiskHypothesisNeedsMoreEvidence, SuggestedNextChecks: []SuggestedNextCheck{NextScopeClarification}}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	it := planItemByHypothesis(p, "rh-out-of-scope-reference")
	if it == nil || it.Status != WebTestPlanStatusBlockedNeedsScope {
		t.Fatalf("expected blocked_needs_scope")
	}
	if it.ProposedTarget != "" {
		t.Fatalf("expected no executable request target")
	}
}

func TestWebTestPlan_AdminAuthPrivatePathsCarryAuthAccessSkills(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-exposure-robots", AffectedPaths: []string{"/admin", "/auth"}, SuggestedExpertSkills: []SuggestedExpertSkill{SkillAuthSecurityReview, SkillAccessControlReview}, SuggestedNextChecks: []SuggestedNextCheck{NextAuthBoundaryReview, NextAccessControlReviewDryRun}}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	it := planItemByHypothesis(p, "rh-exposure-robots")
	if it == nil || !hasExpertSkill(it.SuggestedExpertSkills, SkillAuthSecurityReview) || !hasExpertSkill(it.SuggestedExpertSkills, SkillAccessControlReview) {
		t.Fatalf("expected auth/access-control planning item")
	}
}

func TestWebTestPlan_PreservesHypothesisEvidenceAndSkills(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-security-headers-gap", EvidenceIDs: []string{"ev-9"}, SuggestedExpertSkills: []SuggestedExpertSkill{SkillHeadersSecurityReview}, Priority: RiskPriorityHigh}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	it := planItemByHypothesis(p, "rh-security-headers-gap")
	if it == nil || len(it.EvidenceIDs) != 1 || it.EvidenceIDs[0] != "ev-9" {
		t.Fatalf("expected evidence id preserved")
	}
	if !hasExpertSkill(it.SuggestedExpertSkills, SkillHeadersSecurityReview) {
		t.Fatalf("expected skills preserved")
	}
	if it.Priority != RiskPriorityHigh {
		t.Fatalf("expected priority copied")
	}
}

func TestWebTestPlan_RequiredApprovalsOnlyForBlockedNeedsApproval(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-security-headers-gap"}, {HypothesisID: "rh-unknown-scope-context", Status: RiskHypothesisNeedsMoreEvidence}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	if len(p.RequiredApprovals) != 0 {
		t.Fatalf("expected no required approvals in planning-only blocked-scope scenarios")
	}
}

func TestWebTestPlan_NoProposedMethodAndManualToolClass(t *testing.T) {
	h := RiskHypothesisSet{SetID: "risk-hypotheses-static-mvp", Hypotheses: []RiskHypothesis{{HypothesisID: "rh-security-headers-gap"}}}
	p := BuildWebTestPlanFromHypotheses(h, nil)
	for _, it := range p.TestPlanItems {
		if it.ProposedMethod != "" {
			t.Fatalf("expected no proposed method by default")
		}
		if it.ToolClass != "manual" {
			t.Fatalf("expected manual tool class by default")
		}
	}
}

func planItemByHypothesis(p WebTestPlan, id string) *WebTestPlanItem {
	for i := range p.TestPlanItems {
		if p.TestPlanItems[i].HypothesisID == id {
			return &p.TestPlanItems[i]
		}
	}
	return nil
}

func hasExpertSkill(skills []SuggestedExpertSkill, s SuggestedExpertSkill) bool {
	for _, v := range skills {
		if v == s {
			return true
		}
	}
	return false
}
