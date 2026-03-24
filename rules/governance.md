# Helm Governance

Rules that all agents must follow. Violations are enforced per rule.

---

## 1. Boundaries

Every agent has a defined mission, inputs, outputs, and success criteria.

- Agents operate within their pipeline position
- Agents cannot modify artifacts owned by other agents
- Cross-scope requests are routed through the orchestrator
- Agents report status and score to session state

**Enforcement: BLOCK**

---

## 2. Bar

Quality is measured against concrete, binary (pass/fail) criteria — not subjective assessment.

- Every agent must achieve >= 90% on its success criteria before handoff
- QA agents (review, verify) require >= 95%
- Criteria must be specific to the current execution, not generic checklists
- Silent correction loops: max 3 iterations before escalating to user
- Agents must not define trivially easy criteria to inflate scores

**Enforcement: BLOCK** — Results below threshold are never presented as final.

---

## 3. Relay

Every agent must produce a handoff document upon completion. Every agent must read the previous handoff upon activation.

### Structure
- **Summary** (max 150 lines): Mission completed, key decisions, artifacts produced, critical context for next agent, self-validation score
- **Deep Context** (optional, max 500 lines): Only when complex discoveries don't fit in the summary

### Reading Order
1. Read summary first
2. Read deep context only if present and relevant
3. Read referenced artifacts
4. Fallback: session state if handoff is missing

**Enforcement: BLOCK** — No handoff = no progress.

---

## 4. Memory

All state is persisted so work survives restarts.

- Session state lives in `.helm/session.yaml` (runtime, not committed)
- System config lives in `helm.yaml` (committed)
- Handoffs live in `.helm/handoffs/`
- Decisions are never lost between sessions

**Enforcement: BLOCK** — Agents must persist state before completion.

---

## 5. Guard

- No destructive operations without explicit user confirmation
- Credentials and secrets are never stored in system files
- SAST scanning is mandatory before deployment (verify agent)
- Critical/high vulnerabilities block deployment
- Generated code must follow OWASP Top 10 guidelines

**Enforcement: BLOCK**

---

## 6. Voice

- Agents communicate exclusively through handoffs and session state
- Direct agent-to-agent communication is not allowed
- Supported languages: English (en-US) and Portuguese (pt-BR)
- User-facing messages use the configured language
- All artifacts, agent definitions, and system docs are in English
- Error messages are constructive: what failed, why, how to fix

**Enforcement: GUIDE**

---

## 7. Modes

Three modes control what agents can do:

| Mode | Pipeline States | Read | Write |
|------|----------------|------|-------|
| **planning** | discover, plan | Entire project | `.helm/` only |
| **build** | build, validate | Entire project | Entire project |
| **deploy** | deploy | Entire project | Entire project + infra |

Transitions:
- planning → build: requires review agent score >= 95%
- build → deploy: requires verify agent APPROVED
- build → planning: when verify finds spec/architecture issues (not code issues)

**Enforcement: BLOCK** — Writes outside permitted scope are rejected.

---

## 8. Profiles

Three profiles control how much confirmation is needed:

| Profile | Behavior |
|---------|----------|
| **explore** | Read-only. Agents analyze and suggest but perform no writes. |
| **guided** | Default. Agents propose actions, user confirms before writes. |
| **autonomous** | Agents execute without confirmation when gate scores >= 95%. |

Always require confirmation regardless of profile:
- Destructive operations (file deletion, force push, DB drops)
- Production deployments
- Backward pipeline transitions

If quality drops below 95% during autonomous execution, downgrade to guided.

**Enforcement: BLOCK**

---

## 9. Conduct

Agents lead the conversation. Users validate and respond.

Agents must:
- Know their mission from handoff + pipeline position
- Drive toward completion proactively
- Ask specific questions when input is needed
- Adapt depth to user expertise level

Agents must not:
- Ask "what do you want me to do?" — they already know
- Wait passively for direction
- Expose internal implementation details

**Enforcement: GUIDE**

---

## 10. North Star

1. **Research before code**: Never build without understanding the problem. The DISCOVER and PLAN phases exist because assumptions kill projects.
2. **Short iterations**: Deliver in small, verifiable increments. Every agent validates before passing forward.
3. **Product over process**: Focus on outcomes. A completed task that doesn't serve the product is waste.
4. **No over-engineering**: Build for current requirements. Three similar lines beat a premature abstraction.

### Anti-Patterns
- Skipping research to jump into code
- Phases longer than 2 weeks
- Ignoring user corrections
- Excessive documentation nobody reads

---

*Helm Governance v0.1.0 — 10 Rules*
*All agents are bound by these rules.*
