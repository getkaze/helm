# Testing Patterns

Reference material for the **verify** and **build** agents. Load when writing or validating tests.

---

## Test Pyramid

| Level | Share | Scope | Speed | When to use |
|-------|-------|-------|-------|-------------|
| Unit | ~80% | Single function/method | Milliseconds | Pure logic, transformations, calculations |
| Integration | ~15% | Cross-boundary (DB, HTTP, queues) | Seconds | Service interactions, repository layer |
| E2E | ~5% | Full system path | Seconds–minutes | Critical user flows only |

**Rule:** If you can test it at a lower level, do. Push tests down the pyramid.

---

## Test Structure

### Arrange-Act-Assert (AAA)

Every test has exactly three sections:

```
1. Arrange — set up inputs and expected state
2. Act — call the function under test
3. Assert — verify the output matches expectations
```

### Go — Table-Driven Tests

```go
func TestParseAmount(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr bool
    }{
        {"valid integer", "100", 100, false},
        {"negative", "-50", -50, false},
        {"empty string", "", 0, true},
        {"overflow", "99999999999999999999", 0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseAmount(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

### PHP — Data Providers

```php
/** @dataProvider amountProvider */
public function testParseAmount(string $input, int $expected): void
{
    $this->assertSame($expected, parseAmount($input));
}

public static function amountProvider(): array
{
    return [
        'valid integer' => ['100', 100],
        'negative'      => ['-50', -50],
        'zero'          => ['0', 0],
    ];
}
```

---

## Naming Convention

Test names describe behavior, not implementation:

```
// BAD — describes implementation
TestUserRepository_FindByID_CallsDatabase

// GOOD — describes behavior
TestFindUser_ReturnsNotFoundError_WhenUserDoesNotExist
```

Pattern: `Test{Action}_{ExpectedResult}_{Condition}`

---

## What to Test

For every function, cover these cases:

| Case | Example |
|------|---------|
| Happy path | Valid input → expected output |
| Empty/zero input | `""`, `0`, `nil`, `[]` |
| Boundary values | Max int, max length, off-by-one |
| Error paths | Invalid input, missing resource, timeout |
| Edge cases | Unicode, special chars, concurrent access |

---

## Prove-It Pattern (Bug Fixes)

When fixing a bug, always follow this sequence:

1. **Write a test that reproduces the bug** — it must FAIL
2. **Confirm** the test fails (proving the bug exists)
3. **Implement the fix**
4. **Run the test** — it must PASS (proving the fix works)
5. **Run the full suite** — no regressions

This creates a permanent regression guard. The bug can never silently return.

---

## Test Behavior, Not Implementation

```
// BAD — tests internal method calls
assert(mock.CalledWith("findByID", 42))

// GOOD — tests observable outcome
user := repo.FindByID(42)
assert(user.Name == "Alice")
```

Tests that assert on method calls break when you refactor internals. Tests that assert on outcomes survive refactors.

---

## Mocking Rules

1. **Only mock at system boundaries** — DB, HTTP clients, external APIs, file system
2. **Prefer real implementations** — use in-memory DB, test containers, or embedded servers
3. **Never mock the thing you're testing**
4. **If mocking requires > 5 lines of setup**, the design may need refactoring

### Go — Interfaces at boundaries

```go
// Define interface where you USE it, not where you implement it
type UserStore interface {
    FindByID(ctx context.Context, id int64) (*User, error)
}

// Test with a fake
type fakeUserStore struct{ users map[int64]*User }
func (f *fakeUserStore) FindByID(_ context.Context, id int64) (*User, error) {
    u, ok := f.users[id]
    if !ok { return nil, ErrNotFound }
    return u, nil
}
```

### PHP — Dependency injection

```php
// Interface at boundary
interface UserRepository {
    public function findById(int $id): ?User;
}

// Fake for tests
class InMemoryUserRepository implements UserRepository {
    private array $users = [];
    public function findById(int $id): ?User {
        return $this->users[$id] ?? null;
    }
}
```

---

## Test Isolation

- Each test must be independent — no shared mutable state between tests
- Tests must pass in any order and in parallel
- Clean up after yourself: reset DB state, remove temp files
- Use `t.Parallel()` (Go) to surface hidden dependencies

---

## Test Speed

| Threshold | Action |
|-----------|--------|
| < 10s total | Healthy — run on every save |
| 10–60s | Acceptable — run before commit |
| > 60s | Slow — split into fast/slow suites, tag slow tests |

Slow tests get skipped. Skipped tests rot. Keep the suite fast.

---

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Testing private methods | Couples tests to internals | Test through public API |
| Shared test fixtures | Hidden dependencies, flaky tests | Inline setup per test |
| Sleep in tests | Slow, flaky, race conditions | Use channels/signals/retries with timeout |
| Asserting on logs | Fragile, implementation-coupled | Return errors, assert on behavior |
| Giant test functions | Hard to debug failures | One concept per test, descriptive names |
| 100% coverage goal | Diminishing returns, tests for getters/setters | Cover behavior, not lines |
