# Persistence Contract (Core Skills Pack)

All SDD/SDR/support skills must return persistence guidance even when no persistence backend is active.

Required output per phase:
- what should be persisted
- suggested Engram key
- suggested OpenCode artifact path
- suggested local artifact path
- session state update
- handoff payload summary (if any)

Suggested Engram keys:
- `sdd/{change-name}/proposal`
- `sdd/{change-name}/design`
- `sdd/{change-name}/tasks`
- `sdd/{change-name}/verify`
- `sdr/{topic-id}/capture`
- `sdr/{topic-id}/research`
- `sdr/{topic-id}/synthesis`
- `sdr/{topic-id}/questions`
- `sdr/{topic-id}/structure`
- `sdr/{topic-id}/validation`
- `sdr/{topic-id}/archive`
