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

func mustReadText(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}
