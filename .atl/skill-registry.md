# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

See `_shared/skill-resolver.md` for the full resolution protocol.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When creating a pull request, opening a PR, or preparing changes for review. | branch-pr | /Users/saf/.config/opencode/skills/branch-pr/SKILL.md |
| When creating a GitHub issue, reporting a bug, or requesting a feature. | issue-creation | /Users/saf/.config/opencode/skills/issue-creation/SKILL.md |
| When writing Go tests, using teatest, or adding test coverage. | go-testing | /Users/saf/.config/opencode/skills/go-testing/SKILL.md |
| When user says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen". | judgment-day | /Users/saf/.config/opencode/skills/judgment-day/SKILL.md |
| when the user passes a raw idea, research note, transcript, or topic to evaluate. | research-bitsentry-init | /Users/saf/.config/opencode/skills/research-bitsentry-init/SKILL.md |
| when research-bitsentry-init returns Green and the note needs drafting. | research-bitsentry-create | /Users/saf/.config/opencode/skills/research-bitsentry-create/SKILL.md |
| when a note is ready for final quality review before publishing. | research-bitsentry-validate | /Users/saf/.config/opencode/skills/research-bitsentry-validate/SKILL.md |
| When user asks to create a new skill, add agent instructions, or document patterns for AI. | skill-creator | /Users/saf/.config/opencode/skills/skill-creator/SKILL.md |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### branch-pr
- Every PR MUST link an approved issue and include exactly one `type:*` label.
- Branch names MUST match `type/description` using lowercase `a-z0-9._-` only.
- Use conventional commits only; no `Co-Authored-By` trailers.
- PR body MUST include linked issue, one PR type, summary bullets, changes table, and test plan.
- Run `shellcheck` on modified shell scripts before opening the PR.
- Wait for automated checks to pass before merge.

### issue-creation
- Search for duplicates before creating a new issue.
- Use the correct issue template; blank issues are disabled.
- New issues start with `status:needs-review`; no PR until a maintainer adds `status:approved`.
- Questions belong in Discussions, not Issues.
- Bug reports MUST include repro steps, expected vs actual behavior, OS, agent, and shell.
- Feature requests MUST describe the problem, proposed solution, and affected area.

### go-testing
- Prefer table-driven tests for functions and multi-case logic.
- Test Bubble Tea state transitions by calling `Model.Update()` directly.
- Use `teatest.NewTestModel()` for end-to-end interactive TUI flows.
- Use golden files for stable view/output assertions.
- Mock side effects behind interfaces; use `t.TempDir()` for file-system tests.
- Cover both success and error paths explicitly.

### judgment-day
- Resolve relevant project skills before launching judges; inject the same standards into all prompts.
- Launch TWO blind judges in parallel; neither knows about the other.
- Treat findings confirmed by both judges as high confidence.
- Classify warnings as `real` vs `theoretical`; only real warnings block approval.
- Round 1 fixes require user confirmation before applying.
- After two fix iterations, escalate to the user before continuing.

### research-bitsentry-init
- Act as devil's advocate; do NOT draft the article.
- Compare ideas against `Estadisticas.md` and `PRD.md` before approving.
- Reject clichés, repeated angles, and low-signal topics.
- Output MUST be exactly: `Status`, `Score`, `Reasoning`, `Similar to`, `Risk`.
- Green means strong fit and new value; Yellow needs sharpening; Red is noise.

### research-bitsentry-create
- Only run after `research-bitsentry-init` returns `Status: Green`.
- Start with the problem or thesis; avoid generic intros.
- Keep markdown dense, clean, and aligned with published BitSentry posts.
- Include required frontmatter fields exactly as defined by the skill.
- If the angle is weak or too generic, sharpen or reframe before finalizing.
- Leave the note ready for `Available/`.

### research-bitsentry-validate
- Audit drafts against Posted, Available, Estadisticas, and PRD sources.
- Detect repetition, cliché, weak angles, and cannibalization.
- Keep low-signal drafts out; quality bar is mandatory.
- Return exactly: `Verdict`, `Score`, `Reasoning`, `Overlap`, `Action`.
- Approve only if the note adds something new, reinforces the brand, and is structurally ready.

### skill-creator
- Create a skill only for reusable, non-trivial patterns; otherwise write docs instead.
- Use `skills/{skill-name}/SKILL.md` with complete frontmatter and trigger text.
- Put reusable templates/schemas in `assets/`; local doc pointers in `references/`.
- Keep critical patterns first, examples minimal, and commands copy-pasteable.
- Do not add keyword sections, long explanations, troubleshooting, or web URLs in references.
- Register each new skill in `AGENTS.md` after creation.

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| — | — | No project convention files detected at init time |

Read the convention files listed above for project-specific patterns and rules. All referenced paths have been extracted — no need to read index files to discover more.
