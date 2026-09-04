## 1. Regression Contract

- [x] 1.1 Add a method-aware test matrix for the temporary initial-setup mutation boundary.
- [x] 1.2 Add an integration regression that reproduces HTTP 401 for clean setup with incomplete remote identity configuration.
- [x] 1.3 Run the focused regression before production changes and record the expected failure.

## 2. Runtime Remediation

- [x] 2.1 Implement the exact setup mutation predicate gated by missing initial configuration.
- [x] 2.2 Serialize bootstrap state evaluation and handler dispatch across concurrent setup requests.
- [x] 2.3 Preserve the permanent public API allowlist and outer trusted-host/cross-site request boundary.

## 3. Evidence and Traceability

- [x] 3.1 Update OpenSpec traceability for SETUP-WIZ-022 and the focused regression tests.
- [x] 3.2 Run focused Go request-boundary and runtime setup tests.
- [x] 3.3 Run the focused setup-wizard browser specification against the source runtime.
- [x] 3.4 Run strict OpenSpec validation and repository diff checks.
- [x] 3.5 Record source evidence and the remaining exact-package replay gate on issue #1946.
