## Why

#1481 has migrated OpenAI / ChatGPT into the provider registry, setup schema,
and health/readiness surfaces. The remaining gap is the runtime boundary that
governed Chat will use to ask a configured assistant provider for a normal turn
without giving that provider ambient access to Cabinet tools, the database, or
the filesystem.

## What Changes

- Define a provider-neutral assistant turn interface for governed Chat planner
  consumption.
- Add an OpenAI runtime adapter that resolves credentials, active auth method,
  default model, and health state from the active profile's integration
  registry/instance and secret boundary.
- Add deterministic fake-adapter support for tests with no network calls.
- Normalize assistant output into structured provider text/metadata that the
  governed skill planner can consume without granting provider-side tool
  execution.
- Classify missing credentials, unhealthy providers, unsupported models,
  timeouts, cancellations, rate limits, transport errors, and provider errors
  into actionable redacted outcomes.
- Keep Anthropic, Google, and any other placeholder assistant providers
  unavailable until separate adapters exist.

## Capabilities

### Modified Capabilities

- `provider-openai-chatgpt-ux`: runtime readiness becomes executable only
  through the active profile's configured OpenAI integration instance.
- `assistant-execution-surfaces`: governed Chat may consume assistant-provider
  text output, but Cabinet skill/tool selection remains local to the governed
  planner.
- `ai-gateway`: assistant turns use a provider-neutral request/response shape
  with deterministic fake-provider tests and redacted diagnostics.

## Impact

- Affected code: `internal/ai`, `internal/app`, `internal/chat`,
  integration registry/instance and profile secret helpers.
- Affected tests: provider adapter Go tests, Chat planner/provider-boundary API
  tests, health/error taxonomy tests, and Windows package smoke where provider
  runtime wiring affects packaged behavior.
- Affected documentation: OpenSpec specs, traceability, and #1481 evidence.
- Related issues: `#1481`, `#1933`, `#1932`, `#1714`.
