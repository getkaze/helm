# Security Checklist

Reference material for the **verify** agent. Load when running SAST (Phase 2).

---

## Input Validation

- Validate all external input at system boundaries (HTTP handlers, CLI args, queue consumers)
- Use allowlists over denylists — reject unknown, don't filter known-bad
- Validate type, length, range, and format before processing
- Never trust client-side validation — re-validate server-side

### Go

- Use `strconv` for type conversion, not manual parsing
- Validate struct fields with tags (`validate:"required,min=1,max=255"`)
- Bound slice/map allocations from user input (`make([]T, 0, min(n, maxAllowed))`)

### PHP

- Use `filter_var()` with appropriate filters (`FILTER_VALIDATE_EMAIL`, `FILTER_VALIDATE_INT`)
- Use strict type comparisons (`===`) — loose `==` causes type juggling vulnerabilities
- Validate with allowlists: `in_array($value, $allowed, true)`

---

## SQL Injection

- **Always** use parameterized queries — never concatenate user input into SQL
- ORMs don't guarantee safety — audit raw query methods

### Go

```go
// SAFE — parameterized
db.QueryRow("SELECT * FROM users WHERE id = $1", id)

// DANGEROUS — concatenation
db.QueryRow("SELECT * FROM users WHERE id = " + id)
```

### PHP

```php
// SAFE — prepared statement
$stmt = $pdo->prepare("SELECT * FROM users WHERE id = :id");
$stmt->execute(['id' => $id]);

// DANGEROUS — concatenation
$pdo->query("SELECT * FROM users WHERE id = " . $id);
```

---

## Command Injection

- Never pass user input to shell commands
- If unavoidable: use explicit argument lists, not string interpolation

### Go

```go
// SAFE — argument list
exec.Command("git", "log", "--oneline", ref)

// DANGEROUS — shell interpolation
exec.Command("sh", "-c", "git log --oneline " + ref)
```

### PHP

```php
// SAFE — escapeshellarg
exec('git log --oneline ' . escapeshellarg($ref));

// DANGEROUS — direct interpolation
exec("git log --oneline $ref");
```

---

## Authentication

- Hash passwords with bcrypt (cost >= 12) or argon2
- Implement rate limiting on login endpoints
- Use constant-time comparison for tokens and secrets
- Session tokens: cryptographically random, >= 256 bits
- Invalidate sessions on password change and logout

### Go

```go
bcrypt.GenerateFromPassword([]byte(password), 14)
subtle.ConstantTimeCompare([]byte(a), []byte(b))
```

### PHP

```php
password_hash($password, PASSWORD_BCRYPT, ['cost' => 14]);
hash_equals($expected, $actual);
```

---

## Authorization

- Check authorization on every endpoint — not just in middleware
- Verify resource ownership: `WHERE id = ? AND owner_id = ?`
- Deny by default — require explicit grants
- Separate authentication (who) from authorization (can they)

---

## Secrets Management

- **Never** commit secrets to version control
- Use environment variables or a secret manager (Vault, AWS SSM, GCP Secret Manager)
- Rotate secrets on suspected compromise
- Audit `.env`, config files, and CI/CD variables for leaked credentials

### Patterns to scan for

```
API_KEY=...
SECRET=...
PASSWORD=...
TOKEN=...
PRIVATE_KEY=...
-----BEGIN RSA PRIVATE KEY-----
-----BEGIN OPENSSH PRIVATE KEY-----
```

---

## Path Traversal

- Reject paths containing `..`, `~`, or null bytes
- Resolve to absolute path and verify it stays within allowed directory
- Never use user input directly in file operations

### Go

```go
clean := filepath.Clean(userPath)
abs := filepath.Join(baseDir, clean)
if !strings.HasPrefix(abs, baseDir) {
    // reject
}
```

### PHP

```php
$real = realpath($baseDir . '/' . $userPath);
if ($real === false || !str_starts_with($real, $baseDir)) {
    // reject
}
```

---

## HTTP Security Headers

Set on all responses:

| Header | Value | Purpose |
|--------|-------|---------|
| `Content-Type` | explicit (e.g., `application/json`) | Prevent MIME sniffing |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-Frame-Options` | `DENY` or `SAMEORIGIN` | Prevent clickjacking |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Force HTTPS |
| `Content-Security-Policy` | restrict sources | Prevent XSS |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Limit referrer leakage |

---

## CORS

- Never use `Access-Control-Allow-Origin: *` with credentials
- Allowlist specific origins
- Restrict methods and headers to what's actually needed
- Validate the `Origin` header server-side

---

## Dependency Security

- Run `go vet` + `govulncheck` (Go) or `composer audit` (PHP) before deployment
- Pin dependency versions — don't use floating ranges in production
- Review changelogs before major version bumps
- Remove unused dependencies

---

## Cryptography

- Use standard libraries — never roll your own crypto
- AES-256-GCM for symmetric encryption
- RSA >= 2048 bits or Ed25519 for asymmetric
- TLS 1.2+ for transport — disable older versions
- Never use MD5 or SHA1 for security purposes

---

## Error Handling

- Return generic error messages to clients — never expose stack traces, query details, or internal paths
- Log full error details server-side with request context
- Use structured logging (JSON) for machine-parseable error trails

### Go

```go
// Client sees: {"error": "not found"}
// Server logs: {"error": "user query failed", "query_id": "...", "cause": "..."}
```

### PHP

```php
// Production: display_errors = Off, log_errors = On
// Never: var_dump($exception) in responses
```

---

## OWASP Top 10 Quick Reference

| # | Vulnerability | Primary Defense |
|---|--------------|-----------------|
| 1 | Broken Access Control | Deny by default, check on every request |
| 2 | Cryptographic Failures | TLS everywhere, strong hashing, no custom crypto |
| 3 | Injection | Parameterized queries, argument lists |
| 4 | Insecure Design | Threat model during architecture phase |
| 5 | Security Misconfiguration | Security headers, audit deps, disable debug |
| 6 | Vulnerable Components | `govulncheck` / `composer audit`, pin versions |
| 7 | Auth Failures | Bcrypt/argon2, rate limiting, session management |
| 8 | Data Integrity Failures | Verify signatures, validate CI/CD pipelines |
| 9 | Logging Failures | Structured logging, monitor auth events |
| 10 | SSRF | Allowlist outbound URLs, validate schemes |
