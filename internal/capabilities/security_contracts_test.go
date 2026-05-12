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
