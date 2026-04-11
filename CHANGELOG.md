# Changelog

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
