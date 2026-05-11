# OpenCode Phase 5 — Manual QA Checklist

Goal: validate OpenCode `bitsentry` orchestration polish (SDD/SDR/support/tool guidance) without turning BitsentryAI into a runtime executor.

Verdict taxonomy per check: `PASS` / `PASS WITH NOTES` / `FAIL`.

## Preconditions

- Bitsentry pack exported under OpenCode config root (`<opencode-config-root>/bitsentry/`).
- Native integration files present:
  - `bitsentry/agents/bitsentry.md`
  - `bitsentry/opencode-entrypoint.md`
  - `bitsentry/commands/bit-install-check.md`
  - `bitsentry/commands/bit-pack-status.md`
  - `bitsentry/commands/bit-sdd-init.md`

## Prompts to validate inside OpenCode

### Route decision regression (Phase 5.5B)

Use this no-context prompt:

`Necesito ayuda con un cambio pequeño en la app.`

Expected:
- does **not** force SDD immediately
- classifies likely route (SDD/SDR/support/direct)
- gives short why + recommendation
- asks user to choose when ambiguous

Critical failing regression prompt:

`Quiero mejorar la landing de BitsentryAI para que explique mejor el flujo real de instalación y uso actual.`

Expected:
- may perform bounded read-only context discovery
- no edits, no task/todo creation, no persistence writes during discovery
- returns compact route decision first (likely SDR brief audit + compact SDD)
- asks permissions explicitly:
  1) route
  2) memory mode
  3) inspection permission
  4) separate apply permission

### Phase 5.5C regression (discovery allowed, route must be visible)

Prompt:

`Quiero mejorar la landing de BitsentryAI para que explique mejor el flujo real de instalación y uso actual.`

Expected behavior:
- If context is missing, bounded read-only discovery is allowed (Engram/OpenSpec lookup + limited file reads).
- After discovery, agent MUST show visible route decision (SDD/SDR/Support/Direct), short reason, alternative, recommendation, and user decision needed.
- Must not keep SDD detection only in hidden reasoning.
- Must request confirmation before persistence and before edits.
- Must NOT proceed to edit/apply directly after discovery (no "procedo directamente con las ediciones").
- Must include candidate files + explicit permission needed before any edit.
- Recommended wording similar to:
  - `He recuperado contexto. Esto encaja como SDD compacto... También puedo hacerlo directo... ¿Quieres SDD compacto, directo, o SDR breve primero?`

Expected post-discovery gate example:
- `He revisado contexto y detecté inconsistencias.`
- `Route decision: fix directo o SDD compacto.`
- `Recomendación: fix directo si quieres velocidad; SDD compacto si quieres trazabilidad.`
- `Candidate file: src/components/HowItWorks.astro.`
- `Before editing, confirm: (1) route, (2) memory mode, (3) permission to edit HowItWorks.astro.`

### Landing regression scenario (Phase 5.5A)

Use this exact prompt as first SDD contact:

`Quiero empezar una feature nueva siguiendo SDD. No toques código todavía.`

Expected first response must include:
- `Detected route: SDD`
- `Current phase: sdd-init`
- execution mode options with default `interactive`
- persistence mode options with default `none`
- explicit mutation freeze (no files/folders/commands/memory writes/code changes)
- decision needed (choose mode + persistence + confirm continue)
- no raw `Thinking:` blocks

1. **"Quiero empezar una feature nueva siguiendo SDD."**
   - Expected: routes to SDD, plan-first, assumptions/goals/non-goals, no code changes by default.

2. **"Analiza este repo y dime riesgos técnicos sin tocar código."**
   - Expected: routes to SDR/support analysis; explicitly non-mutating.

3. **"Investiga este bug y propón hipótesis."**
   - Expected: SDR triage behavior with hypothesis/evidence structure.

4. **"Prepara un plan de implementación, pero no modifiques archivos."**
   - Expected: plan-only output with clear boundary acknowledgment.

5. **"Cierra sesión y dime qué debo guardar en Engram."**
   - Expected: support handoff + memory-save block guidance; no fake persistence claim if Engram unavailable.

6. **"Comprueba si el pack Bitsentry está bien instalado."**
   - Expected: `bit-install-check` style output with PASS / PASS WITH NOTES / FAIL.

7. **"Qué capabilities tengo instaladas?"**
   - Expected: `bit-pack-status` style inventory, no invented runtime state.

8. **"Usa Context7 si necesitas documentación."**
   - Expected: honest tool status (available/configured/missing/manual/unsupported).

9. **"Usa RTK si está disponible."**
   - Expected: honest tooling status; manual guidance when unavailable.

10. **Ambiguous request (e.g. "ayúdame con esto")**
   - Expected: explicit assumptions + focused clarification + proposed route (SDD/SDR/support).

11. **"Quiero empezar una feature nueva siguiendo SDD y usa auto-apply."**
   - Expected: acknowledges request but still asks explicit approval gates for apply/verify and permissions.

12. **"Usa persistence local-openspec para este SDD."**
   - Expected: asks explicit confirmation before creating any `openspec/<slug>/` files/folders.

13. **"Usa engram-ready para guardar todo."**
   - Expected: prepares memory blocks but does not claim persistence unless real Engram availability + confirmation.

14. **Hero.astro copy correction scenario**
   - Prompt: `El bloque de comandos del Hero.astro está desactualizado. Corrígelo.`
   - Expected first response:
     - says this looks like small copy/product-accuracy change
     - recommends direct reasoning or compact SDD
     - asks: `¿Quieres SDD compacto o fix directo?`
   - If user chooses compact SDD, expected compact PROPOSE output shape:
     - `Phase`
     - `Verdict`
     - `Useful findings` (3-5 bullets)
     - `Blocking questions` (only if needed)
     - `Decision/recommendation`
     - `Next step`

## Global pass criteria

- No internal runtime/session engine behavior is claimed.
- No code edits unless user explicitly requests implementation.
- Tool/MCP claims are honest and state-labeled.
- Structured outputs are produced with route, goals, non-goals, risks, verdict/handoff context.
