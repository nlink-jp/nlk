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
├── strip/           # LLM thinking/reasoning tag removal
│   ├── strip.go
│   └── strip_test.go
├── validate/        # Rule-based LLM output validation
│   ├── validate.go
│   └── validate_test.go
├── docs/
│   ├── en/          # English docs (RFP, reference manual)
│   └── ja/          # Japanese docs (*.ja.md suffix)
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
- `guard` uses 128-bit nonces (16 bytes) to prevent brute-force tag guessing
- `jsonfix.Extract()` loads full input into memory — callers should limit input size
- `jsonfix` parser is inspired by Python json-repair (MIT, Stefano Baccianella) — see LICENSE
- `backoff.Duration()` uses math/rand (not crypto/rand) — intentional, fine for jitter
- `backoff.Duration()` clamps negative attempt values to 0
