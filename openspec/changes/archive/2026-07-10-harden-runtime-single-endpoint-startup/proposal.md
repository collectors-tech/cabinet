## Why

Cabinet now avoids some accidental duplicate launches, but the runtime behavior is still underspecified when a requested host/port is already serving a healthy Cabinet instance. We need a clear, testable product contract so normal launches prefer reuse of the active endpoint while intentional parallel-instance workflows remain available.

## What Changes

- Clarify runtime startup behavior when the requested endpoint is already serving Cabinet.
- Define default single-endpoint reuse semantics for normal launches.
- Add an explicit restart option that stops the currently running Cabinet on the requested endpoint and starts the new process on that same endpoint.
- Define the explicit parallel-instance override path so intentional multi-instance runs still work.
- Tighten startup diagnostics so attach vs restart vs fallback decisions are observable and deterministic.
- Align launcher/runtime/demo scripts with the runtime singleton decision instead of relying on ad hoc port-in-use checks.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `general/runtime-core`: refine startup behavior so a healthy requested Cabinet endpoint is reused by default instead of starting a second runtime or silently port-falling back.
- `general/runtime-multi-instance`: clarify that parallel-instance behavior is an explicit opt-in path and must not weaken the default singleton/startup attach contract.

## Impact

- `cmd/cabinet` startup decision flow and attach diagnostics
- runtime restart/termination path for existing Cabinet processes
- runtime port negotiation / endpoint health decision logic
- launch scripts such as `scripts/runtime/start-demo2.ps1`
- runtime tests covering attach, fallback, and explicit parallel mode
- OpenSpec runtime requirements for launcher behavior
