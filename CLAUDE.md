# Helm

AI agent orchestration for backend development.

## How to Use
Type `/helm` to start. The orchestrator reads session state and routes to the correct agent.

## Structure
- `agents/` — 12 agent definitions (11 specialized + orchestrator)
- `rules/governance.md` — 10 governance rules
- `schemas/` — JSON schemas for session and config validation
- `.helm/` — Runtime state (not committed)
