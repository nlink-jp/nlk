# AGENTS.md — nlk

## Summary

Lightweight Go library providing LLM utility packages for nlink-jp projects.
Toolbox of pure functions for the code surrounding LLM API calls.
Zero external dependencies.

## Build & Test

```bash
go test ./...              # Run all tests
go test ./... -cover       # With coverage
go test ./... -v           # Verbose
```

No Makefile — this is a library, not a binary.

## Project Structure

```
nlk/
├── guard/           # Nonce-tagged XML wrapping (prompt injection defense)
│   ├── guard.go
│   └── guard_test.go
├── jsonfix/         # Recursive descent JSON parser with repair
│   ├── jsonfix.go   # Public API (Extract, ExtractTo)
│   ├── parser.go    # Recursive descent parser
│   └── jsonfix_test.go
├── backoff/         # Exponential backoff duration calculation
│   ├── backoff.go
│   └── backoff_test.go
├── validate/        # Rule-based LLM output validation
│   ├── validate.go
│   └── validate_test.go
├── docs/
│   └── design/      # RFP, reference manual
├── go.mod
├── README.md
├── README.ja.md
├── CHANGELOG.md
└── LICENSE
```

## Design Principles

- Each package is independent — no cross-package imports
- Zero external dependencies (standard library only)
- Pure functions, stateless, no side effects
- Does NOT abstract LLM API calls

## Gotchas

- `guard.NewTag()` panics if crypto/rand fails (should never happen in practice)
- `jsonfix.Extract()` uses greedy regex — picks the largest JSON match
- `backoff.Duration()` uses math/rand (not crypto/rand) — fine for jitter timing
