# RFP: nlk

> Generated: 2026-04-11
> Status: Implemented (v0.3.2)
>
> **Note:** This document records the initial planning discussion. The final
> implementation differs in several areas due to design decisions made during
> development. See "Implementation Notes" at the end for a summary of changes.

## 1. Problem Statement

Multiple nlink-jp LLM applications (mail-analyzer, gem-cli, ai-ir2, news-collector, etc.) each independently implement common LLM SDK peripheral functions (prompt injection guards, LLM output JSON repair, output validation, retry backoff calculation, etc.), resulting in duplicated code across projects. These will be consolidated into a lightweight shared library usable from both Go and Python. The library does NOT abstract LLM calls themselves — it focuses exclusively on the code surrounding those calls. It is a toolbox, not a framework.

## 2. Functional Specification

### Commands / API Surface

A Go library (`github.com/nlink-jp/nlk`), not a CLI tool. Provides 5 packages:

| Package | Function | Nature |
|---------|----------|--------|
| `guard` | Nonce-tagged XML wrapping for prompt injection defense | Pure data transformation |
| `prompt` | Prompt template construction | Pure data transformation |
| `jsonfix` | Malformed JSON extraction, repair, and parsing | Pure data transformation |
| `validate` | Validation function execution framework (rules defined by app) | Pure validation |
| `backoff` | Exponential backoff duration calculation with jitter | Pure computation |

All packages are pure functions, stateless. Each package is independently usable.

### Input / Output

Each package is used via function calls. No stdin/stdout involved.

```go
// guard: wrap untrusted data
guarded := guard.Wrap(untrustedData)

// jsonfix: repair malformed JSON
parsed, err := jsonfix.Parse(rawLLMOutput)

// backoff: calculate wait duration
time.Sleep(backoff.Duration(attempt))

// validate: run app-defined rules
err := validate.Run(result, myRules...)

// prompt: build prompt from template
text := prompt.Build(template, vars)
```

### Configuration

No config files or environment variables needed. Everything is controlled via function arguments.

### External Dependencies

None. All packages are implemented using Go standard library only.

## 3. Design Decisions

### Tech Stack

- **Go**: Initial implementation language. Many nlink-jp LLM tools are written in Go.
- **Python**: Follow-up in a separate repository (`nlk-py`)

### Design Principles

- **Toolbox, not framework** — does not dictate call flow. Apps compose freely.
- **No LLM API abstraction** — does not unify Vertex AI and OpenAI-compatible API differences. LLM calls are the app's responsibility.
- **Zero external dependencies** — supply chain attack mitigation. Go standard library only.
- **Pure functions, stateless** — no side effects in any package.

### Relationship with Existing Tools

- `guard`: extracted from gem-cli and mail-analyzer injection guards
- `jsonfix`: ported from json-filter's JSON extraction/repair logic
- `backoff`: extracted from mail-analyzer and gem-cli retry logic
- `validate`: generalized from mail-analyzer's output validation patterns
- `prompt`: commonized from prompt construction patterns across tools

### Position in LLM Application Workflow

```
  User Input / External Data
        |
        v
  +---------+
  |  guard   |  Pre-processing: wrap untrusted data
  +----+----+
       v
  +---------+
  |  prompt  |  Pre-processing: assemble prompt
  +----+----+
       v
    LLM API      App-specific (Vertex AI / OpenAI-compatible / local)
       |         backoff.Duration() for wait calculation; loop is app's responsibility
       v
  +---------+
  | jsonfix  |  Post-processing: extract/repair JSON from raw response
  +----+----+
       v
  +----------+
  | validate  |  Post-processing: validate with app-defined rules
  +----+-----+
       v
  Application Logic
```

### Out of Scope

- LLM API call abstraction
- Pipeline/workflow engine
- Retry loop control (backoff only calculates wait duration)
- Python version (follow-up in separate repository)
- JSON Schema validation (validate only provides app-defined function execution framework)

## 4. Development Plan

### Phase 1: Core

- `guard` package implementation + tests
- `jsonfix` package implementation + tests (ported from json-filter)
- `backoff` package implementation + tests
- High test coverage mandatory as a foundation library

### Phase 2: Features

- `prompt` package implementation + tests
- `validate` package implementation + tests
- Migrate one existing tool (e.g., mail-analyzer) to nlk for validation

### Phase 3: Release

- Documentation (README.md, docs/ja/README.md, CHANGELOG.md)
- AGENTS.md
- Go module publication (tagging)

### Review Units

Each Phase can be reviewed independently.

## 5. Required API Scopes / Permissions

None. The library itself does not call external APIs.

## 6. Series Placement

Series: **lib-series** (new)
Reason: A cross-series foundation library. Not appropriate for any existing series. A dedicated library series is created.

## 7. External Platform Constraints

None. A pure Go library with no external platform dependencies.

---

## Discussion Log

1. **Problem recognition**: Many LLM apps independently implement similar SDK wrappers and injection guards
2. **Scope decision**: No LLM call abstraction. Unifying Vertex AI and OpenAI-compatible APIs would recreate existing heavy SDKs. Focus on "surrounding" code only
3. **Language**: Target Go/Python both, but start with Go
4. **retry consideration**: Initially considered a retry module → proposed pipeline → recognized framework risk → settled on backoff duration calculation only
5. **Dependency policy**: All packages zero external dependencies. Critical for supply chain attack mitigation. If dependencies become necessary, declare explicitly
6. **validate design**: Not JSON Schema-compliant generic validation, but app-provided validation function execution (Option C). Maintains zero dependencies
7. **Naming**: `nlk` (abbreviation of nlink-jp). Prevents namespace collision while remaining concise
8. **Series**: Foundation library spanning all series — does not belong to any existing series; `lib-series` created

---

## Implementation Notes

The following summarizes changes made during implementation.
The RFP body is preserved as-is as a record of planning-phase discussion.

### Package Composition Changes

| RFP Plan | Final Implementation | Reason |
|----------|---------------------|--------|
| `prompt` | **Not implemented** | Survey of existing tools showed guard.Expand() + fmt.Sprintf is sufficient. A dedicated package was excessive |
| — | `strip` **added** | Need for LLM thinking/reasoning tag removal. Provided as independent package rather than inside jsonfix (zero cross-package dependency principle) |

Final packages: `guard`, `jsonfix`, `strip`, `backoff`, `validate`

### API Name Changes

| RFP | Implementation |
|-----|---------------|
| `jsonfix.Parse()` | `jsonfix.Extract()` — clarifies "extraction + repair" intent |
| `validate.Run(result, rules...)` | `validate.Run(rules...)` — result is captured by each rule's closure |

### jsonfix Implementation Approach Change

RFP stated "ported from json-filter", but the final implementation is a recursive descent parser written from scratch in Go, informed by Python json-repair (MIT, Stefano Baccianella). Regex-based approach had ReDoS risk and limited repair coverage.

### Security Hardening

- guard: nonce size increased from 4 bytes (32-bit) → 16 bytes (128-bit). Prevents brute-force tag guessing
- backoff: negative attempt values clamped to 0
- jsonfix: memory usage note added to documentation for large inputs

### Workflow Diagram (Final)

```
  User Input / External Data
        |
        v
  +---------+
  |  guard   |  Pre-processing: wrap untrusted data
  +----+----+
       v
    LLM API      App-specific (Vertex AI / OpenAI-compatible / local)
       |         backoff.Duration() for wait calculation; loop is app's responsibility
       v
  +---------+
  |  strip   |  Post-processing: remove thinking/reasoning tags (local LLM)
  +----+----+
       v
  +---------+
  | jsonfix  |  Post-processing: extract/repair JSON from raw response
  +----+----+
       v
  +----------+
  | validate  |  Post-processing: validate with app-defined rules
  +----+-----+
       v
  Application Logic
```
