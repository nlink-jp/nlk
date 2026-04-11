# Changelog

## v0.5.0 (2026-04-12)

### Security

- **Breaking:** guard: `Wrap()` now returns `(string, error)` — returns `ErrTagCollision` if input data contains the tag name (defense-in-depth)
- guard: document that Tag must be generated per LLM call (never reuse across turns)
- jsonfix: add JSON smuggling risk note to `Extract()` doc
- strip: add input size warning to `ThinkTags()` doc

## v0.4.0 (2026-04-11)

### Documentation

- Add Example test functions for all 5 packages (visible in `go doc`)
- Add `examples/workflow/` — complete guard → LLM → strip → jsonfix → validate pipeline
- Restructure `docs/` into `docs/en/` and `docs/ja/`
- Add English RFP document
- Add Implementation Notes to RFP (plan vs reality)
- Sync all documentation with code after security review

### Security

- guard: increase nonce from 4 bytes (32-bit) to 16 bytes (128-bit)
- backoff: clamp negative attempt values to 0
- backoff: document intentional use of math/rand
- jsonfix: add memory usage note for large inputs

### Fixes

- jsonfix: handle unescaped double quotes in strings (look-ahead heuristic)
- jsonfix: handle double-escaped JSON strings (`{\"key\": \"value\"}`)
- validate: fix doc comment
- json-repair MIT license attribution added to LICENSE and parser.go

## v0.3.2 (2026-04-11)

### Features

- `jsonfix`: handle double-escaped JSON strings (`{\"key\": \"value\"}`)

## v0.3.1 (2026-04-11)

### Fixes

- `jsonfix`: handle unescaped double quotes inside JSON strings using look-ahead heuristic
- jsonfix coverage improved to 94.2%

### Tests

- Added tests for unescaped inner quotes, truncated input, EOF edge cases, unquoted values with code fences

## v0.3.0 (2026-04-11)

### New Packages

- `strip` — LLM thinking/reasoning tag removal
  - `<think>...</think>` (DeepSeek R1, Qwen QwQ/3, Phi-4, most OSS models)
  - `<thinking>`, `<reasoning>`, `<reflection>` variants
  - `<|channel>thought...<channel|>` (Gemma 4)
  - Case-insensitive matching, unclosed tags, empty tags
  - 100% test coverage

### Documentation

- Reference manual updated with strip package (en/ja)
- Fixed mojibake in Japanese reference manual

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
