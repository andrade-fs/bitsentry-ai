package securityweb

import (
	"net/url"
	"strings"
	"testing"
)

func TestDiscoveryPlanMVPRequestsAndConstraints(t *testing.T) {
	adapter := NewOfflineAdapter(DefaultPolicyEvaluator{}, DefaultDryRunPlanner{}, NewDefaultEvidenceRecorder(DefaultRedactor{}))
	ctx := baseContext()
	ctx.ExecutionMode = ExecutionModePlanningOnly

	plan := adapter.BuildDiscoveryPlan(ctx, "example.com")

	if plan.WouldExecute {
		t.Fatalf("would_execute must always be false")
	}

	required := map[string]bool{
		"HEAD https://example.com/":                              false,
		"GET https://example.com/":                               false,
		"GET https://example.com/robots.txt":                     false,
		"GET https://example.com/sitemap.xml":                    false,
		"GET https://example.com/.well-known/security.txt":       false,
	}

	for _, item := range plan.Items {
		sig := string(item.Request.Method) + " " + item.Request.URL
		if _, ok := required[sig]; ok {
			required[sig] = true
		}

		if item.Request.Method != MethodGET && item.Request.Method != MethodHEAD {
			t.Fatalf("planned method must be GET/HEAD, got %s", item.Request.Method)
		}

		u, err := url.Parse(item.Request.URL)
		if err != nil {
			t.Fatalf("invalid planned url %q: %v", item.Request.URL, err)
		}
		if !strings.EqualFold(u.Hostname(), "example.com") {
			t.Fatalf("planned url must remain in-scope, got host=%q", u.Hostname())
		}
		if u.RawQuery != "" {
			t.Fatalf("planner must not add dynamic query params: %q", item.Request.URL)
		}

		if item.EvidenceTemplate.PlannedRequestRef != item.Request.RequestRef {
			t.Fatalf("evidence template must be linked to request ref")
		}
		if item.EvidenceTemplate.EvidenceID == "" {
			t.Fatalf("evidence template must have stable id")
		}
	}

	for sig, found := range required {
		if !found {
			t.Fatalf("missing required discovery request: %s", sig)
		}
	}

	for _, item := range plan.Items {
		if item.Request.URL == "https://example.com/favicon.ico" || item.Request.URL == "https://example.com/.well-known/change-password" {
			t.Fatalf("optional path must not be in MVP: %s", item.Request.URL)
		}
	}
}

func TestDiscoveryPlanOutOfScopeTargetDenied(t *testing.T) {
	adapter := NewOfflineAdapter(DefaultPolicyEvaluator{}, DefaultDryRunPlanner{}, NewDefaultEvidenceRecorder(DefaultRedactor{}))
	ctx := baseContext()
	ctx.ExecutionMode = ExecutionModePlanningOnly

	plan := adapter.BuildDiscoveryPlan(ctx, "other.example")

	if len(plan.Items) == 0 {
		t.Fatalf("expected planned items")
	}

	for _, item := range plan.Items {
		if !hasViolation(item.PolicyViolations, "target_out_of_scope") {
			t.Fatalf("out-of-scope target must be denied per planned request")
		}
	}
}

func TestDiscoveryPlanWouldExecuteFalseInPlanningOnlyAndDryRun(t *testing.T) {
	adapter := NewOfflineAdapter(DefaultPolicyEvaluator{}, DefaultDryRunPlanner{}, NewDefaultEvidenceRecorder(DefaultRedactor{}))

	planning := baseContext()
	planning.ExecutionMode = ExecutionModePlanningOnly
	if adapter.BuildDiscoveryPlan(planning, "example.com").WouldExecute {
		t.Fatalf("planning_only must keep would_execute=false")
	}

	dryRun := baseContext()
	dryRun.ExecutionMode = ExecutionModeDryRun
	if adapter.BuildDiscoveryPlan(dryRun, "example.com").WouldExecute {
		t.Fatalf("dry_run must keep would_execute=false")
	}
}
