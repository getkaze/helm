# Launch Checklist

Reference material for the **ship** agent. Load on activation.

---

## Pre-Launch Checklist

Complete all applicable items before deployment.

### 1. Code Quality

- [ ] All tests pass (`go test ./...` / `phpunit`)
- [ ] Build succeeds with no warnings
- [ ] Linter passes (`golangci-lint` / `phpstan` / `phpcs`)
- [ ] No `TODO`, `FIXME`, or `HACK` without linked issue
- [ ] No `fmt.Println` / `var_dump` / `print_r` debug output
- [ ] Error handling covers expected failure modes
- [ ] Code reviewed by verify agent (APPROVED)

### 2. Security

- [ ] No secrets in code or config files
- [ ] Vulnerability scan clean (`govulncheck` / `composer audit`)
- [ ] Input validation at all system boundaries
- [ ] Auth/authz on every endpoint
- [ ] Security headers configured
- [ ] CORS restricted to specific origins
- [ ] Rate limiting on public/auth endpoints

### 3. Performance

- [ ] No N+1 queries
- [ ] All new queries use indexes (`EXPLAIN`)
- [ ] Pagination on list endpoints
- [ ] Response times within targets (P95)
- [ ] Connection pools configured
- [ ] No unbounded allocations from user input

### 4. Data

- [ ] Migrations tested (up AND down)
- [ ] Migrations are backward-compatible with current code
- [ ] Seed data updated if applicable
- [ ] Backup verified before destructive migrations

### 5. Infrastructure

- [ ] Environment variables documented and set
- [ ] Health check endpoint responds
- [ ] Logging configured (structured, appropriate levels)
- [ ] Error reporting connected (Sentry, Bugsnag, etc.)
- [ ] DNS / TLS / certificates valid
- [ ] CI/CD pipeline green

### 6. Documentation

- [ ] README updated (if setup changed)
- [ ] API documentation current (new/changed endpoints)
- [ ] Changelog updated
- [ ] Deployment notes written (what changed, how to rollback)

---

## Rollout Strategy

### Staged Rollout Sequence

| Stage | Scope | Duration | Gate |
|-------|-------|----------|------|
| 1. Staging | Full test suite, smoke tests | Until green | All tests pass |
| 2. Deploy (flag off) | Production infra, feature inactive | 1 hour | No errors from deploy itself |
| 3. Internal | Team only | 24 hours | No new errors |
| 4. Canary | 5% of traffic | 24–48 hours | Metrics within thresholds |
| 5. Gradual | 25% → 50% → 100% | 24h per step | Metrics within thresholds |
| 6. Full | 100%, monitor | 1 week | Stable |

### Decision Thresholds

| Metric | Green (proceed) | Yellow (hold) | Red (rollback) |
|--------|-----------------|---------------|-----------------|
| Error rate | < 10% above baseline | 10–100% above | > 2x baseline |
| P95 latency | < 20% above baseline | 20–50% above | > 50% above |
| Failed requests | < 0.1% | 0.1–1% | > 1% |
| Business metrics | Neutral or positive | Decline < 5% | Decline > 5% |

**Any RED metric = immediate rollback. No debate.**

---

## Rollback Plan

Every deployment must have a documented rollback path:

1. **How to rollback**: exact command or pipeline trigger
2. **What data is affected**: migrations, cache, queues
3. **Who to notify**: on-call, stakeholders, downstream services
4. **Recovery time target**: how long rollback takes

### Rollback Patterns

| Change Type | Rollback Method |
|-------------|----------------|
| Feature flag | Turn off flag |
| Code-only (no migration) | Deploy previous version |
| Code + migration (backward-compatible) | Deploy previous version (migration stays) |
| Code + migration (breaking) | Run down migration, then deploy previous version |
| Infrastructure | Terraform/Pulumi revert, DNS failover |

**Rule:** If you can't describe the rollback in one sentence, the deployment is too risky for a single release. Split it.

---

## Feature Flag Lifecycle

1. **Deploy with flag OFF** — code in production, inactive
2. **Enable for team** — internal testing in production
3. **Gradual rollout** — 5% → 25% → 50% → 100%
4. **Monitor at each stage** — errors, latency, business metrics
5. **Clean up** — remove flag and dead code within 2 weeks of full rollout

**Rule:** A feature flag that lives longer than 30 days without full rollout needs a decision: ship it or kill it.

---

## Post-Launch Verification

After reaching 100% rollout:

1. [ ] All monitoring dashboards reviewed
2. [ ] Error rate stable (not trending up)
3. [ ] Latency stable (not trending up)
4. [ ] No unexpected resource consumption (CPU, memory, disk)
5. [ ] Feature flags cleaned up
6. [ ] Deployment notes archived
