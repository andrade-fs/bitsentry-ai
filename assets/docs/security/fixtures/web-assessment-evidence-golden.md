# Evidence Contract Golden Fixture

- Evidence ID: web-assessment-evidence-001
- Source: synthetic log excerpt
- Target / URL: https://synthetic-target.example.invalid/login
- In scope confirmation: yes (authorized staging target)
- Authorization reference: AUTH-GOLDEN-001
- Request method: GET
- Request purpose: verify session security headers
- Tool class: Single-request verification
- Intensity: low
- Timestamp / testing window: 2026-05-10T10:12:00Z (authorized window)
- Result summary: response returned 200 with missing Secure cookie flag
- Relevant headers / status / behavior: status=200; Set-Cookie missing Secure
- Safety notes: no exploit attempts, no destructive actions
- Redactions: user identifiers redacted
- Linked finding IDs: WA-F-001
- Limitations: single endpoint sample, no production target interaction
