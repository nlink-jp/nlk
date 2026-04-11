# Changelog

## v0.2.0 (2026-04-11)

### New Packages

- `validate` — rule-based LLM output validation (OneOf, Range, MaxLen, NotEmpty, Custom)

### Improvements

- `jsonfix` — complete rewrite with recursive descent parser (inspired by Python's json-repair)
  - Single quotes → double quotes
  - Trailing commas removal
  - Unquoted keys
  - Missing commas insertion
  - Comments removal (// /* */ #)
  - TRUE/FALSE/NULL/None normalization
  - Python tuples → arrays
  - Ellipsis removal
  - Leading/trailing dot in numbers
  - Underscore separators in numbers
  - Hex escapes → unicode escapes

### Documentation

- Reference manual (en/ja) with full API documentation

### Test Coverage

- backoff: 100%, guard: 90%, jsonfix: 88.6%, validate: 100%

## v0.1.0 (2026-04-11)

Initial release.

### Packages

- `guard` — nonce-tagged XML wrapping for prompt injection defense
- `jsonfix` — JSON extraction and repair from LLM output
- `backoff` — exponential backoff duration calculation with jitter

### Design

- Zero external dependencies (standard library only)
- Pure functions, stateless
- Each package is independent
