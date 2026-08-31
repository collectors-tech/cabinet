## Why

A clean Cabinet runtime renders the setup wizard before authentication, but its
setup mutation requests are currently rejected by the session boundary with
HTTP 401 when an incomplete remote-identity configuration activates that
boundary. This makes the documented **Use Defaults**, manual storage validation
and config-import paths dead-end on first launch and blocks packaged onboarding
acceptance for #1946 and 1.0 GA.

## What Changes

- Permit only the setup completion, config import and storage validation
  mutations to cross the authentication boundary while the runtime still
  requires initial setup.
- Preserve trusted-host, same-origin and request-boundary protections for the
  bootstrap mutation.
- Remove the special bootstrap exception once setup is complete so normal local
  or remote session policy applies again.
- Add regression coverage for the pre-auth success and post-setup rejection
  boundaries before changing production behavior.
- Record OpenSpec and traceability evidence for the corrected first-run contract.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `setup-wizard-first-run`: Clarify that setup completion is an
  unauthenticated bootstrap mutation only while configuration is missing, and
  that subsequent requests return to the configured session policy.

## Impact

- Runtime request/session boundary and setup-completion handler in
  `internal/app`.
- Go request-boundary and setup API regression tests.
- Focused first-run UI/package evidence under issue #1946.
- OpenSpec delta and traceability for the setup wizard capability.
