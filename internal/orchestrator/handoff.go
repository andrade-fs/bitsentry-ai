package orchestrator

import "strings"

func RenderBriefMarkdown(brief Brief) string {
	var b strings.Builder
	b.WriteString("# Brief\n\n")
	b.WriteString("## Intent\n")
	b.WriteString("- Original: " + brief.IntentOriginal + "\n")
	b.WriteString("- Normalized: " + brief.IntentNormalized + "\n\n")
	b.WriteString("## Flow\n")
	b.WriteString("- " + brief.Flow + "\n\n")
	b.WriteString("## Constraints\n")
	if len(brief.Constraints) == 0 {
		b.WriteString("- (none)\n\n")
	} else {
		for _, c := range brief.Constraints {
			b.WriteString("- " + c + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("## Non-goals\n")
	for _, ng := range brief.NonGoals {
		b.WriteString("- " + ng + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

func RenderHandoffMarkdown(h Handoff) string {
	var b strings.Builder
	b.WriteString("# Handoff\n\n")
	b.WriteString("- Session ID: " + h.SessionID + "\n")
	b.WriteString("- From: " + h.From + "\n")
	b.WriteString("- To: " + h.To + "\n")
	b.WriteString("- Reason: " + h.Reason + "\n\n")
	b.WriteString("## Next Steps\n")
	for _, s := range h.NextSteps {
		b.WriteString("- " + s + "\n")
	}
	b.WriteString("\n## Safety Notes\n")
	for _, n := range h.SafetyNotes {
		b.WriteString("- " + n + "\n")
	}
	b.WriteString("\n")
	return b.String()
}
