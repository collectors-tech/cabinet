## 1. Contract

- [x] 1.1 Define the assistant-provider runtime boundary, provider-neutral turn
  interface, OpenAI adapter scope, fake-adapter test path, and non-goal
  boundaries for #1481.
- [x] 1.2 Add OpenSpec traceability rows for the assistant-provider runtime
  requirements and validation evidence.

## 2. Runtime adapter

- [x] 2.1 Add a provider-neutral assistant turn request/response type used by
  governed Chat planning.
- [x] 2.2 Add a deterministic fake assistant provider adapter for tests without
  network calls.
- [x] 2.3 Implement the OpenAI adapter through the active configured
  integration instance and profile secret boundary.
- [x] 2.4 Resolve model/default options from schema-driven provider
  configuration and reject unsupported models with redacted guidance.

## 3. Safety and diagnostics

- [x] 3.1 Enforce timeout, cancellation, retry/rate-limit, and transport error
  classification with bounded tests.
- [x] 3.2 Keep provider credentials inside Cabinet secret storage and redact
  logs, errors, workflow evidence, and health diagnostics.
- [x] 3.3 Prove providers cannot call Cabinet skills, database, filesystem, or
  app-control tools directly; #1933 owns governed tool selection/dispatch.
- [x] 3.4 Keep Anthropic/Google placeholders unavailable until separate adapters
  exist.

## 4. Evidence

- [x] 4.1 Add focused Go tests for fake adapter, OpenAI adapter setup/readiness,
  normal turn completion, and redacted error taxonomy.
- [x] 4.2 Add Chat/API coverage showing governed Chat consumes provider output
  without provider-side Cabinet tool execution.
- [x] 4.3 Run strict OpenSpec validation, focused Go tests, touched UI/API
  validation where relevant, Windows package smoke if runtime wiring affects
  packaged behavior, and `git diff --check`.
