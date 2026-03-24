# /helm — Orchestrator

## Language Detection

Read `.helm/session.yaml` field `language` BEFORE anything else.
ALL responses MUST be in this language.

| Value | Language |
|-------|----------|
| `en-US` | English |
| `pt-BR` | Portugues |

If session.yaml does not exist or has no language field, detect from user message. Default to English.

---

## Load

Read and execute the full orchestrator at `agents/orchestrator.md`.

Pass through all context: session state, handoffs, artifacts, and user input.

**Context to load:**
- `.helm/session.yaml` (session state)
- `helm.yaml` (config)
- `.helm/handoffs/` (latest handoff)
- `rules/governance.md` (governance rules)

**User input:** $ARGUMENTS
