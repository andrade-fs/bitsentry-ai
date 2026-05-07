# Skill Loading Contract

Orchestrators should load shared contracts first:

1. result-envelope
2. persistence-contract
3. handoff-contract
4. backend conventions (engram/opencode)

Then pass compact instructions to subskills (fresh context per phase).

Orchestrators are delegate-only by default.
