## 1. Startup Decision Contract

- [x] 1.1 Add/adjust runtime startup tests that distinguish healthy Cabinet endpoint reuse from generic port-occupied fallback behavior.
- [x] 1.2 Add explicit-parallel-mode tests proving requested-endpoint reuse is bypassed only when parallel startup is intentionally enabled.
- [x] 1.3 Add restart-mode tests proving a healthy requested Cabinet endpoint can be stopped and replaced on the same port.
- [x] 1.4 Verify attach/restart diagnostics clearly identify same-data-dir attach, requested-endpoint reuse, requested-endpoint restart, and fallback-port outcomes.

## 2. Runtime / Launcher Implementation

- [x] 2.1 Refine startup decision ordering so Cabinet checks for an existing healthy requested endpoint before port fallback.
- [x] 2.2 Keep non-Cabinet port occupation on the existing deterministic fallback path.
- [x] 2.3 Add an explicit restart startup option that resolves the active Cabinet PID, stops the old process, waits for the requested port to clear, and starts the replacement runtime on the same endpoint.
- [x] 2.4 Align launcher/browser-open behavior so reuse or restart of an existing requested endpoint does not start a second server process.

## 3. Parallel / Tooling Alignment

- [x] 3.1 Make explicit parallel-mode handling consistent across CLI/runtime code and helper launch scripts.
- [x] 3.2 Update review/demo lane startup helpers to expose restart behavior and rely on the runtime singleton contract instead of ad hoc duplicate-start checks where appropriate.
- [x] 3.3 Validate multi-instance/stress tooling still uses isolated data roots and does not regress under the stricter default singleton behavior.

## 4. Verification and Rollout

- [x] 4.1 Run targeted runtime tests covering same-data-dir attach, requested-endpoint reuse, requested-endpoint restart, non-Cabinet fallback, and explicit parallel override.
- [x] 4.2 Run targeted launcher/script verification for demo/startup workflows affected by the new attach/restart decision path.
- [x] 4.3 Update implementation notes/checkpoints with the final attach/restart/fallback behavior and example diagnostics.
