# Performance Checklist

Reference material for the **verify** agent. Load when reviewing performance (Phase 3).

---

## Measure First

Never optimize without profiling. Guesses about bottlenecks are wrong more often than right.

### Go

```bash
# CPU profile
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profile
go test -bench=. -memprofile=mem.prof
go tool pprof -alloc_space mem.prof

# Trace
go test -trace=trace.out
go tool trace trace.out
```

### PHP

```bash
# Xdebug profiling
php -d xdebug.mode=profile -d xdebug.output_dir=./profiles script.php

# Blackfire (if available)
blackfire run php script.php
```

---

## Database

| Issue | Detection | Fix |
|-------|-----------|-----|
| N+1 queries | Query count grows linearly with result set | Eager load / JOIN / batch query |
| Missing indexes | Slow query log, `EXPLAIN` shows full scan | Add index on filtered/sorted columns |
| No pagination | Response size grows unbounded | Cursor or offset pagination, enforce `LIMIT` |
| Connection exhaustion | Timeouts under load | Connection pooling, bound max connections |
| Unbound `IN` clauses | Query with thousands of parameters | Batch into chunks, use temp table |
| Missing `WHERE` on `UPDATE`/`DELETE` | Affects all rows | Always filter, use transactions |

### Query Performance Targets

| Metric | Target |
|--------|--------|
| Simple lookup (indexed) | < 5ms |
| Complex JOIN | < 50ms |
| Aggregation/report | < 500ms |
| Total queries per request | < 10 |

### Go — Connection Pool

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)
```

### PHP — Persistent Connections

```php
// PDO with persistent connection
$pdo = new PDO($dsn, $user, $pass, [
    PDO::ATTR_PERSISTENT => true,
]);
```

---

## API Response Times

| Endpoint Type | P50 Target | P95 Target | P99 Target |
|--------------|------------|------------|------------|
| Health check | < 5ms | < 10ms | < 50ms |
| CRUD read | < 50ms | < 200ms | < 500ms |
| CRUD write | < 100ms | < 300ms | < 1s |
| Search/filter | < 200ms | < 500ms | < 2s |
| Report/export | < 1s | < 5s | < 10s |

If P95 exceeds target, investigate before shipping.

---

## Memory

### Go

| Issue | Detection | Fix |
|-------|-----------|-----|
| Goroutine leak | `runtime.NumGoroutine()` grows | Use `context.Context` with cancel/timeout |
| Unbounded buffer | Memory grows with input size | Stream processing, fixed-size buffers |
| Large allocations in hot paths | `pprof -alloc_objects` | Sync.Pool, pre-allocate, reuse slices |
| Holding references | Objects never GC'd | Nil pointers after use, weak refs |

### PHP

| Issue | Detection | Fix |
|-------|-----------|-----|
| Memory limit exceeded | `Allowed memory size exhausted` | Process in chunks, use generators |
| Loading full dataset into array | `memory_get_peak_usage()` | Use cursors, generators, `yield` |
| Circular references | Memory not freed between requests | `gc_collect_cycles()`, break cycles |

---

## Concurrency

### Go

| Issue | Detection | Fix |
|-------|-----------|-----|
| Race condition | `go test -race` | Mutex, channels, or atomic operations |
| Unbounded goroutines | Goroutine count under load | Worker pool with semaphore |
| Deadlock | Hang under load | Lock ordering, `context.WithTimeout` |
| Channel misuse | Goroutine leak, panic | Always close from sender, use `select` with default |

```go
// Worker pool pattern
sem := make(chan struct{}, maxWorkers)
for _, item := range items {
    sem <- struct{}{}
    go func(it Item) {
        defer func() { <-sem }()
        process(it)
    }(item)
}
```

### PHP

| Issue | Detection | Fix |
|-------|-----------|-----|
| Blocking I/O | Slow response under concurrent load | Queue async jobs, use non-blocking where possible |
| Long-running request | Request timeout | Move to background job, return early |
| Session locking | Serial requests per user | `session_write_close()` early, or use token-based sessions |

---

## Caching

| Layer | Use case | TTL |
|-------|----------|-----|
| Application cache (Redis/Memcached) | Computed results, API responses | Minutes to hours |
| Query cache | Expensive DB queries | Seconds to minutes |
| HTTP cache headers | Static/semi-static responses | `Cache-Control`, `ETag` |

**Rules:**
- Cache at the highest level possible
- Always set TTL — no infinite caches
- Invalidate on write, don't rely on TTL alone for consistency
- Monitor hit rate — below 80% means the cache isn't helping

---

## Serialization

| Issue | Fix |
|-------|-----|
| Serializing entire DB row when client needs 3 fields | Use DTOs / response structs with only needed fields |
| JSON encoding in hot loop | Reuse encoder, consider binary format (protobuf, msgpack) |
| Large response payloads | Paginate, compress (gzip/brotli), use `fields` query param |

---

## Logging & Observability

| Issue | Impact | Fix |
|-------|--------|-----|
| Logging in hot paths | Latency spike | Log on error only, sample debug logs |
| Synchronous log writes | Blocks request | Async log writer, buffer + flush |
| No request tracing | Can't diagnose slow requests | Add trace ID to all logs and responses |
| Missing metrics | Blind to degradation | Instrument: request count, latency histogram, error rate |

---

## Pre-Ship Performance Gate

Before approving for deployment, verify:

1. [ ] No N+1 queries (check query count per request)
2. [ ] All queries use indexes (run `EXPLAIN` on new queries)
3. [ ] Response times within targets (P95)
4. [ ] No goroutine/memory leaks under sustained load (Go)
5. [ ] No unbounded allocations from user input
6. [ ] Connection pools configured with limits
7. [ ] Pagination on all list endpoints
8. [ ] Race detector passes (Go: `go test -race`)
