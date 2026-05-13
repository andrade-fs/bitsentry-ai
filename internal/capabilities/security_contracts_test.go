package capabilities

import (
	"os"
	"strings"
	"testing"
)

func TestSecurityFindingsSkillMinimumContractTokens(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/security-findings/SKILL.md")

	requiredTokens := []string{
		"- ID",
		"- Title",
		"- Severity",
		"- Confidence",
		"- Category",
		"- Affected files",
		"- Affected component",
		"- Evidence",
		"- Impact",
		"- Likelihood",
		"- Remediation",
		"- Verification",
		"- References / Notes",
	}

	for _, token := range requiredTokens {
		if !strings.Contains(content, token) {
			t.Fatalf("security-findings contract missing token %q", token)
		}
	}
}

func TestWebRequestAdapterDesignHas77BImplementedMappingAnchor(t *testing.T) {
	content := mustReadText(t, "../../docs/design/web-request-adapter.md")
	required := []string{
		"### 12) 7.7B implemented contracts mapping",
		"internal/securityweb/",
		"sin network execution",
	}
	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("web-request-adapter design missing 7.7B mapping token %q", token)
		}
	}
}

func TestControlledHTTPExecutorDesignHas79BPEPAndNoPolicyTransportAnchors(t *testing.T) {
	content := mustReadText(t, "../../docs/design/controlled-http-executor.md")
	required := []string{
		"### 10) 7.9B Real HTTP transport skeleton (httptest-only)",
		"OfflineControlledExecutor remains the single Policy Enforcement Point",
		"transport does not enforce scope/approval/method/tool/rate/budget policy",
		"redirect not followed by default",
		"Location captured",
		"no full response stored by default",
	}
	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("controlled-http-executor design missing 7.9B anchor token %q", token)
		}
	}
}

func TestWebRequestAdapterHas79BTransportBoundaryAnchor(t *testing.T) {
	content := mustReadText(t, "../../docs/design/web-request-adapter.md")
	required := []string{
		"### 17) 7.9B minimal real transport note",
		"internal/securitywebhttp",
		"PEP remains unique in internal/securityweb",
	}
	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("web-request-adapter design missing 7.9B boundary token %q", token)
		}
	}
}

func TestSecurityFindingsSkillSeverityAndConfidenceEnums(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/security-findings/SKILL.md")

	severityValues := []string{
		"- Critical",
		"- High",
		"- Medium",
		"- Low",
		"- Informational",
	}
	for _, v := range severityValues {
		if !strings.Contains(content, v) {
			t.Fatalf("security-findings contract missing severity value %q", v)
		}
	}

	confidenceValues := []string{"- High", "- Medium", "- Low"}
	for _, v := range confidenceValues {
		if !strings.Contains(content, v) {
			t.Fatalf("security-findings contract missing confidence value %q", v)
		}
	}
}

func TestSecurityReportSkillRequiredSections(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/security-report/SKILL.md")

	requiredSections := []string{
		"- Title",
		"- Executive Summary",
		"- Scope",
		"- Methodology",
		"- Repository / Application Context",
		"- Risk Summary",
		"- Findings",
		"- Evidence",
		"- Remediation Plan",
		"- Verification Steps",
		"- Assumptions and Limitations",
		"- Next Steps",
		"- Appendix",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Fatalf("security-report contract missing required section %q", section)
		}
	}
}

func TestSecurityFindingsToReportExplicitHandoff(t *testing.T) {
	findings := mustReadText(t, "../../assets/skills/security/security-findings/SKILL.md")
	report := mustReadText(t, "../../assets/skills/security/security-report/SKILL.md")

	handoff := "security-findings -> security-report"
	if !strings.Contains(findings, handoff) {
		t.Fatalf("security-findings missing explicit handoff %q", handoff)
	}
	if !strings.Contains(report, handoff) {
		t.Fatalf("security-report missing explicit handoff consumption %q", handoff)
	}
}

func TestSecurityFindingsOfficialTaxonomy(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/security-findings/SKILL.md")

	categories := []string{
		"- Authentication",
		"- Authorization",
		"- Session Management",
		"- Input Validation",
		"- Injection",
		"- Cross-Site Scripting",
		"- Server-Side Request Forgery",
		"- File Upload",
		"- Secrets Exposure",
		"- Cryptography",
		"- Dependency Risk",
		"- GraphQL Security",
		"- Business Logic",
		"- Configuration",
		"- Logging / Monitoring",
		"- Error Handling",
		"- Data Exposure",
		"- Supply Chain",
		"- Informational",
	}

	for _, c := range categories {
		if !strings.Contains(content, c) {
			t.Fatalf("security-findings missing official category %q", c)
		}
	}
}

func TestSecurityFindingsCalibrationAndRulesAnchors(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/security-findings/SKILL.md")

	requiredAnchors := []string{
		"Impact × Likelihood",
		"Critical: impacto muy alto + likelihood alta o exposición clara con riesgo sistémico.",
		"High: impacto alto con likelihood razonable, o impacto crítico con evidencia parcial.",
		"Medium: impacto moderado o explotación condicionada.",
		"Low: impacto limitado, alcance reducido o mitigaciones claras.",
		"Informational: observación útil sin impacto explotable confirmado.",
		"High: evidencia directa en código/config no secreta, flujo claro y baja ambigüedad.",
		"Medium: patrón razonable con algunas asunciones.",
		"Low: señal débil, contexto incompleto o requiere verificación manual.",
		"## Deduplication Rules",
		"## Evidence Grouping Rules",
		"## Assumptions / Limitations Rules",
		"## Skill → Category Mapping Contract",
	}

	for _, token := range requiredAnchors {
		if !strings.Contains(content, token) {
			t.Fatalf("security-findings missing calibration/rule anchor %q", token)
		}
	}
}

func TestSecurityFindingsSkillCategoryMappingAnchors(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/security-findings/SKILL.md")

	mappings := []string{
		"auth-security-review -> primary: Authentication | secondary: Authorization, Session Management",
		"jwt-review -> primary: Session Management | secondary: Authentication, Cryptography",
		"graphql-security-review -> primary: GraphQL Security | secondary: Authorization, Injection, Data Exposure",
		"xss-review -> primary: Cross-Site Scripting | secondary: Input Validation",
		"file-upload-review -> primary: File Upload | secondary: Input Validation, Configuration",
		"ssrf-review -> primary: Server-Side Request Forgery | secondary: Input Validation, Configuration",
		"secrets-review -> primary: Secrets Exposure | secondary: Configuration, Supply Chain",
		"dependency-risk-review -> primary: Dependency Risk | secondary: Supply Chain, Configuration",
	}

	for _, m := range mappings {
		if !strings.Contains(content, m) {
			t.Fatalf("security-findings missing skill->category mapping %q", m)
		}
	}
}

func TestSecurityReportConsumesTaxonomyInRiskSummaryAndFindings(t *testing.T) {
	report := mustReadText(t, "../../assets/skills/security/security-report/SKILL.md")

	requiredTokens := []string{
		"## Taxonomy Consumption Contract",
		"### Risk Summary (Required Consumption)",
		"Aggregate by Category (official enum), Severity, and Confidence.",
		"### Findings (Required Consumption)",
		"Preserve each finding Category exactly as provided by `security-findings` taxonomy.",
		"Preserve Severity and Confidence enums exactly.",
		"Preserve assumptions/limitations relevant to exploitability confidence.",
	}

	for _, token := range requiredTokens {
		if !strings.Contains(report, token) {
			t.Fatalf("security-report missing taxonomy consumption token %q", token)
		}
	}
}

func TestSecurityDocsFixturesAndExamplesExist(t *testing.T) {
	paths := []string{
		"../../assets/docs/security/README.md",
		"../../assets/docs/security/examples/findings-example.md",
		"../../assets/docs/security/examples/report-example.md",
		"../../assets/docs/security/fixtures/findings-golden.md",
		"../../assets/docs/security/fixtures/report-golden.md",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected docs artifact %s: %v", p, err)
		}
	}
}

func TestSecurityFindingsFixturesContractAndCoverage(t *testing.T) {
	files := []string{
		"../../assets/docs/security/examples/findings-example.md",
		"../../assets/docs/security/fixtures/findings-golden.md",
	}

	requiredTokens := []string{
		"- ID:",
		"- Title:",
		"- Severity:",
		"- Confidence:",
		"- Category:",
		"- Affected files:",
		"- Affected component:",
		"- Evidence:",
		"- Impact:",
		"- Likelihood:",
		"- Remediation:",
		"- Verification:",
		"- References / Notes:",
	}

	for _, file := range files {
		content := mustReadText(t, file)
		for _, token := range requiredTokens {
			if !strings.Contains(content, token) {
				t.Fatalf("%s missing required finding contract token %q", file, token)
			}
		}
	}

	golden := mustReadText(t, "../../assets/docs/security/fixtures/findings-golden.md")
	for _, sev := range []string{"Critical", "High", "Medium", "Low", "Informational"} {
		if !strings.Contains(golden, "- Severity: "+sev) {
			t.Fatalf("findings-golden missing severity coverage for %q", sev)
		}
	}

	example := mustReadText(t, "../../assets/docs/security/examples/findings-example.md")
	for _, conf := range []string{"High", "Medium", "Low"} {
		if !strings.Contains(example, "- "+conf) {
			t.Fatalf("findings-example missing confidence enum value %q", conf)
		}
	}

	requiredCategories := []string{
		"Authorization",
		"Session Management",
		"Server-Side Request Forgery",
		"File Upload",
		"Informational",
	}
	for _, c := range requiredCategories {
		if !strings.Contains(golden, "- Category: "+c) {
			t.Fatalf("findings-golden missing realistic taxonomy category %q", c)
		}
	}

	for _, token := range []string{"Deduplication", "Evidence Grouping", "Assumptions / Limitations"} {
		if !strings.Contains(golden, token) && !strings.Contains(example, token) {
			t.Fatalf("findings docs missing rules/examples for %q", token)
		}
	}
}

func TestSecurityReportGoldenRequiredSectionsOrder(t *testing.T) {
	content := mustReadText(t, "../../assets/docs/security/fixtures/report-golden.md")

	sections := []string{
		"# Title",
		"## Executive Summary",
		"## Scope",
		"## Methodology",
		"## Repository / Application Context",
		"## Risk Summary",
		"## Findings",
		"## Evidence",
		"## Remediation Plan",
		"## Verification Steps",
		"## Assumptions and Limitations",
		"## Next Steps",
		"## Appendix",
	}

	last := -1
	for _, s := range sections {
		idx := strings.Index(content, s)
		if idx == -1 {
			t.Fatalf("report-golden missing required section %q", s)
		}
		if idx <= last {
			t.Fatalf("report-golden section out of order at %q", s)
		}
		last = idx
	}
}

func TestSecurityDocsGuardrailsPresent(t *testing.T) {
	content := mustReadText(t, "../../assets/docs/security/README.md")
	guardrails := []string{
		"read-only first",
		"no .env access",
		"no secrets",
		"no exploit execution",
		"no external target testing",
		"no destructive actions",
		"no MCP credential mutation",
		"no runtime flow execution",
		"no autonomous mode",
		"no edits by default",
		"OpenCode-first",
		"CLI debug/plumbing only",
		"agent.bitsentry.permission.edit = deny",
	}

	for _, g := range guardrails {
		if !strings.Contains(content, g) {
			t.Fatalf("security docs README missing guardrail %q", g)
		}
	}
}

func mustReadText(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func TestWebAssessmentRequestsCanonicalToolingPolicyAnchors(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/web-assessment-requests/SKILL.md")

	requiredAnchors := []string{
		"Tooling Policy / Command Safety",
		"no execution by default",
		"explicit approval per request",
		"authorized target required",
		"exact scope required",
		"allowed tools required",
		"prohibited actions required",
		"rate limits required",
		"stop conditions required",
		"evidence logging required",
		"no exploit execution",
		"no destructive actions",
		"no DoS/load testing",
		"no credential attacks",
		"no mass scanning",
		"no out-of-scope scanning",
		"no secrets exposure",
		"no brute force",
		"no password spraying",
		"no aggressive fuzzing",
		"no exfiltration",
	}

	for _, token := range requiredAnchors {
		if !strings.Contains(content, token) {
			t.Fatalf("web-assessment-requests missing mandatory policy anchor %q", token)
		}
	}
}

func TestWebAssessmentRequestsToolClassesContract(t *testing.T) {
	content := mustReadText(t, "../../assets/skills/security/web-assessment-requests/SKILL.md")

	toolClasses := []string{
		"Passive inspection",
		"Single-request verification",
		"Low-noise mapping",
		"Authenticated test with provided test credentials",
		"Prohibited / requires separate explicit approval",
	}

	for _, token := range toolClasses {
		if !strings.Contains(content, token) {
			t.Fatalf("web-assessment-requests missing tool class %q", token)
		}
	}
}

func TestWebAssessmentSkillsPolicySummaryAndCanonicalReference(t *testing.T) {
	paths := []string{
		"../../assets/skills/security/web-assessment-recon-plan/SKILL.md",
		"../../assets/skills/security/web-assessment-map/SKILL.md",
		"../../assets/skills/security/web-assessment-test-plan/SKILL.md",
		"../../assets/skills/security/web-assessment-findings/SKILL.md",
		"../../assets/skills/security/web-assessment-report/SKILL.md",
	}

	for _, p := range paths {
		content := mustReadText(t, p)
		for _, token := range []string{
			"Tooling Policy / Command Safety",
			"no execution by default",
			"security/web-assessment-requests",
		} {
			if !strings.Contains(content, token) {
				t.Fatalf("%s missing policy summary/canonical reference token %q", p, token)
			}
		}
	}
}

func TestWebAssessmentSkillsRequiredSemanticsAnchors(t *testing.T) {
	checks := map[string]string{
		"../../assets/skills/security/web-assessment-recon-plan/SKILL.md": "web-assessment-recon-plan: planifica, no ejecuta.",
		"../../assets/skills/security/web-assessment-map/SKILL.md":        "web-assessment-map: mapea desde evidencia autorizada, no escanea fuera de scope.",
		"../../assets/skills/security/web-assessment-test-plan/SKILL.md":  "web-assessment-test-plan: define intensidad, rate limits, stop conditions y logging plan.",
		"../../assets/skills/security/web-assessment-findings/SKILL.md":   "web-assessment-findings: findings solo con evidencia autorizada.",
		"../../assets/skills/security/web-assessment-report/SKILL.md":     "web-assessment-report: reporta límites, autorización, intensidad y evidencia.",
	}

	for path, token := range checks {
		content := mustReadText(t, path)
		if !strings.Contains(content, token) {
			t.Fatalf("%s missing required semantics anchor %q", path, token)
		}
	}
}

func TestWebAssessmentDocsArtifactsExist(t *testing.T) {
	paths := []string{
		"../../assets/docs/security/examples/web-assessment-report-example.md",
		"../../assets/docs/security/fixtures/web-assessment-report-golden.md",
		"../../assets/docs/security/fixtures/web-assessment-evidence-golden.md",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected web-assessment docs artifact %s: %v", p, err)
		}
	}
}

func TestWebAssessmentEvidenceContractAnchors(t *testing.T) {
	content := mustReadText(t, "../../assets/docs/security/fixtures/web-assessment-evidence-golden.md")
	required := []string{
		"- Evidence ID:",
		"- Source:",
		"- Target / URL:",
		"- In scope confirmation:",
		"- Authorization reference:",
		"- Request method:",
		"- Request purpose:",
		"- Tool class:",
		"- Intensity:",
		"- Timestamp / testing window:",
		"- Result summary:",
		"- Relevant headers / status / behavior:",
		"- Safety notes:",
		"- Redactions:",
		"- Linked finding IDs:",
		"- Limitations:",
	}

	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("web-assessment evidence fixture missing required token %q", token)
		}
	}
}

func TestWebAssessmentReportContractAnchorsAndOrder(t *testing.T) {
	content := mustReadText(t, "../../assets/docs/security/fixtures/web-assessment-report-golden.md")
	sections := []string{
		"# Title",
		"## Authorization Summary",
		"## Scope",
		"## Out of Scope",
		"## Methodology",
		"## Tooling and Intensity",
		"## Request / Evidence Log",
		"## Risk Summary",
		"## Findings",
		"## Remediation Plan",
		"## Verification Steps",
		"## Assumptions and Limitations",
		"## Next Steps",
		"## Appendix",
	}

	last := -1
	for _, s := range sections {
		idx := strings.Index(content, s)
		if idx == -1 {
			t.Fatalf("web-assessment report golden missing section %q", s)
		}
		if idx <= last {
			t.Fatalf("web-assessment report golden section out of order at %q", s)
		}
		last = idx
	}
}

func TestWebAssessmentSkillsTraceabilityAnchor(t *testing.T) {
	trace := "authorization → scope → request/evidence → finding → report"
	for _, p := range []string{
		"../../assets/skills/security/web-assessment-findings/SKILL.md",
		"../../assets/skills/security/web-assessment-report/SKILL.md",
	} {
		content := mustReadText(t, p)
		if !strings.Contains(content, trace) {
			t.Fatalf("%s missing traceability anchor %q", p, trace)
		}
	}
}

func TestWebAssessmentContractsInSkills(t *testing.T) {
	findings := mustReadText(t, "../../assets/skills/security/web-assessment-findings/SKILL.md")
	for _, token := range []string{
		"## Evidence Contract (minimum, exact anchors)",
		"- Evidence ID",
		"- Target / URL",
		"- Relevant headers / status / behavior",
		"- Linked finding IDs",
	} {
		if !strings.Contains(findings, token) {
			t.Fatalf("web-assessment-findings missing token %q", token)
		}
	}

	report := mustReadText(t, "../../assets/skills/security/web-assessment-report/SKILL.md")
	for _, token := range []string{
		"## Report Contract (minimum, exact anchors)",
		"- Authorization Summary",
		"- Out of Scope",
		"- Request / Evidence Log",
		"- Impact Chain",
		"- Retest",
		"- Assumptions and Limitations",
	} {
		if !strings.Contains(report, token) {
			t.Fatalf("web-assessment-report missing token %q", token)
		}
	}
}

func TestWebAssessmentPhase76LifecycleAnchors(t *testing.T) {
	recon := mustReadText(t, "../../assets/skills/security/web-assessment-recon-plan/SKILL.md")
	for _, token := range []string{
		"Assessment Session Context",
		"Intent",
		"Scope",
		"Discovery",
		"Surface Ranking",
		"Risk Hypotheses",
		"planning_only",
		"dry_run",
		"execute_approved",
		"retest",
		"recon / hunt / validate / report",
		"evidence log",
	} {
		if !strings.Contains(recon, token) {
			t.Fatalf("web-assessment-recon-plan missing phase 7.6 anchor %q", token)
		}
	}

	mapSkill := mustReadText(t, "../../assets/skills/security/web-assessment-map/SKILL.md")
	for _, token := range []string{"Mapping", "surface", "Surface Ranking", "Risk Hypotheses"} {
		if !strings.Contains(mapSkill, token) {
			t.Fatalf("web-assessment-map missing phase 7.6 anchor %q", token)
		}
	}

	findings := mustReadText(t, "../../assets/skills/security/web-assessment-findings/SKILL.md")
	for _, token := range []string{
		"Retest Contract (subphase inside findings/report)",
		"Retest plan",
		"Retest status",
		"fixed / partially fixed / still vulnerable / not tested",
		"retest evidence",
		"impact chaining",
	} {
		if !strings.Contains(strings.ToLower(findings), strings.ToLower(token)) {
			t.Fatalf("web-assessment-findings missing phase 7.6 anchor %q", token)
		}
	}

	report := mustReadText(t, "../../assets/skills/security/web-assessment-report/SKILL.md")
	for _, token := range []string{"Impact Chain", "Retest", "lifecycle anchors", "impact chaining", "evidence log"} {
		if !strings.Contains(report, token) {
			t.Fatalf("web-assessment-report missing phase 7.6 anchor %q", token)
		}
	}
}

func TestWebAssessmentReadmeContractSectionAndGuardrails(t *testing.T) {
	content := mustReadText(t, "../../assets/docs/security/README.md")
	for _, token := range []string{
		"## web-assessment contracts",
		"source-security-review",
		"web-assessment-*",
		"web-assessment-report-golden.md",
		"web-assessment-evidence-golden.md",
		"no runtime",
		"no flow execution",
		"no tooling operativo",
		"no target testing vivo",
		"no exploits",
		"no DoS/load testing",
		"no credential attacks",
		"no secrets exposure",
		"evidence must redact sensitive data",
		"no MCP credential mutation",
		"no autonomous mode",
		"OpenCode-first",
		"CLI debug/plumbing only",
		"agent.bitsentry.permission.edit = deny",
	} {
		if !strings.Contains(content, token) {
			t.Fatalf("security README missing web-assessment contract token %q", token)
		}
	}
}

func TestWebRequestAdapterPhase77CanonicalDesignAnchors(t *testing.T) {
	content := mustReadText(t, "../../docs/design/web-request-adapter.md")

	anchors := []string{
		"Purpose / Non-goals",
		"Relationship to web-assessment lifecycle",
		"Contract sources",
		"web-assessment.yaml",
		"web-assessment-requests/SKILL.md",
		"Evidence/Report contracts",
		"PolicyEvaluator",
		"DryRunPlanner",
		"EvidenceRecorder",
		"Redactor",
		"Executor (future, out of scope 7.7)",
		"ControlledWebRequestAdapter",
		"PlannedRequest",
		"AssessmentSessionContext",
		"PolicyDecision",
		"EvidenceEntry",
		"PolicyViolation",
		"planning_only",
		"dry_run",
		"execute_approved",
		"retest",
		"scope validation",
		"scheme validation",
		"redirect policy",
		"method policy",
		"rate limits",
		"request budget",
		"timeout",
		"max response size",
		"stop conditions",
		"redaction",
		"evidence IDs",
		"ErrMissingAuthorization",
		"ErrScopeViolation",
		"ErrOutOfScopeRedirect",
		"ErrMissingRateLimit",
		"ErrMissingStopConditions",
		"ErrToolClassNotAllowed",
		"ErrExecutionModeDenied",
		"ErrMissingEvidencePlan",
		"Evidence model + Markdown template",
		"Layered test strategy",
		"7.7B offline Go stubs",
		"7.8 Passive Discovery MVP",
		"7.9 Controlled Crawler MVP",
		"7.10 Safe Check Modules",
	}

	for _, token := range anchors {
		if !strings.Contains(content, token) {
			t.Fatalf("web-request-adapter design missing anchor %q", token)
		}
	}
}

func TestWebRequestAdapterPhase77ContractualSafetyAnchors(t *testing.T) {
	content := strings.ToLower(mustReadText(t, "../../docs/design/web-request-adapter.md"))

	anchors := []string{
		"no network requests",
		"planning_only",
		"dry_run",
		"execute_approved",
		"retest",
		"policyevaluator",
		"dryrunplanner",
		"evidencerecorder",
		"redactor",
		"authorized target required",
		"explicit approval per request",
		"get/head default",
		"no secrets in logs",
		"out-of-scope redirects denied",
		"request budget",
		"rate limits",
		"timeout",
		"max response size",
		"stop conditions",
	}

	for _, token := range anchors {
		if !strings.Contains(content, token) {
			t.Fatalf("web-request-adapter design missing contractual safety anchor %q", token)
		}
	}
}

func TestControlledHTTPExecutorPhase78BContractAnchors(t *testing.T) {
	content := mustReadText(t, "../../docs/design/controlled-http-executor.md")

	required := []string{
		"Controlled HTTP Executor",
		"execute_approved only",
		"per-request approval",
		"approval_id",
		"approved_request_id",
		"approved_method",
		"approved_url",
		"expires_at",
		"approval_text_or_hash",
		"follow_redirects=false",
		"no redirects followed by default",
		"out-of-scope redirect requires new approval",
		"GET/HEAD only first MVP",
		"one request at a time",
		"no POST",
		"no payloads",
		"no crawler",
		"no scanner",
		"no background execution",
		"body_preview_redacted",
		"headers_redacted",
		"no full response stored by default",
		"FakeTransport only for future tests",
		"no real network in 7.8B",
	}

	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("controlled-http-executor design missing 7.8B contract token %q", token)
		}
	}
}


func TestControlledHTTPExecutorPhase78DHardeningAnchors(t *testing.T) {
	content := mustReadText(t, "../../docs/design/controlled-http-executor.md")
	required := []string{
		"7.8D Hardening Addendum",
		"approval_scope_missing",
		"approval_execution_mode_missing",
		"approval_tool_class_missing",
		"approval_intensity_missing",
		"approval_actor_missing",
		"approval_proof_missing",
		"approval_exceeds_context_limits",
		"redirect_location_invalid",
		"redirect_out_of_scope",
		"request_ref -> approval_id -> evidence_id -> execution_result",
		"no real network in 7.8D",
	}
	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("controlled-http-executor design missing 7.8D hardening token %q", token)
		}
	}
}

func TestControlledHTTPExecutorPhase79ABoundaryAnchors(t *testing.T) {
	content := mustReadText(t, "../../docs/design/controlled-http-executor.md")
	required := []string{
		"internal/securitywebhttp",
		"core offline-safe",
		"Policy Enforcement Point",
		"OfflineControlledExecutor is the single Policy Enforcement Point",
		"real transport must not decide policy",
		"deny before transport",
		"transport receives only policy-approved requests",
		"execute_approved only",
		"per-request approval",
		"GET/HEAD only first MVP",
		"follow_redirects=false",
		"no external network tests",
		"httptest only for future real transport tests",
		"no redirects followed by default",
		"no full response stored by default",
		"body preview capped and redacted",
	}
	for _, token := range required {
		if !strings.Contains(content, token) {
			t.Fatalf("controlled-http-executor design missing 7.9A boundary token %q", token)
		}
	}
}
