## Context

The setup wizard is intentionally rendered before sign-in when `cabinet.json`
is missing. Its three mutation endpoints—completion, config import and storage
validation—are nevertheless evaluated by the same local/ZITADEL session
middleware as an established workspace. An incomplete ZITADEL environment can
therefore activate remote-session enforcement while setup status still requires
the wizard, producing a 401 dead end.

The outer request boundary already validates trusted hosts, origins and
cross-site mutation metadata. The change must preserve that protection and must
not turn setup endpoints into permanently public APIs.

## Goals / Non-Goals

**Goals:**

- Allow the exact setup mutation endpoints to reach their handlers without a
  session while `cabinet.json` is missing.
- Apply the same behavior consistently to completion, import and storage
  validation.
- Stop applying the bootstrap exception immediately after setup exists.
- Preserve the existing request-boundary, method and payload validation.
- Cover the incomplete-ZITADEL first-run failure with a red-before-green Go test.

**Non-Goals:**

- Redesign onboarding steps or authentication modes.
- Change established-workspace authorization policy.
- Expand the permanent public API allowlist.
- Claim packaged acceptance before the focused package replay is performed.

## Decisions

### Use a dynamic bootstrap predicate before session middleware

Add a small predicate that returns true only when all of these conditions hold:

1. `cabinet.json` is missing according to the runtime setup state;
2. the request method is `POST`; and
3. the path is exactly `/api/runtime/setup-complete`,
   `/api/runtime/setup-import` or `/api/runtime/setup-storage-validate`.

The protected handler evaluates this predicate before ZITADEL or local unlock
requirements and dispatches directly to the existing route handler only for
that bootstrap state. The outer request boundary continues to run first.

Adding the routes to `isPublicAPIRequest` was rejected because it would make
them public after setup. Requiring a session token from the wizard was rejected
because no valid identity ceremony exists before initial configuration.

### Serialize bootstrap-state evaluation and handling

A mutex scoped to the application instance covers the bootstrap-state
re-evaluation and route handling. This prevents concurrent completion/import
requests from both retaining the temporary exception after one creates the
configuration file. Requests that arrive after setup exists continue through
the normal session policy.

### Keep route-level validation unchanged

The existing handlers remain responsible for payload validation and config
writes. A bootstrap request can therefore receive normal `200` or deterministic
`400` handler responses, but it cannot bypass JSON, auth-mode, path or storage
checks.

## Risks / Trade-offs

- **First-run mutations are temporarily reachable before identity setup** → The
  exception is restricted by exact method/path, setup state and the existing
  trusted-host/origin boundary, then disappears as soon as config exists.
- **Filesystem setup state can change between checks** → Hold an app-scoped
  mutex across the final state check and handler execution.
- **Future setup endpoints could be forgotten** → Keep an explicit allowlist and
  a method-aware regression matrix; additions require a contract change.
- **Source success could be mistaken for package proof** → Keep packaged replay
  as an unchecked #1946/GA gate after implementation.

## Migration Plan

No persisted schema or data migration is required. Deploy the runtime change,
run focused Go and setup-wizard checks, then reproduce clean-start onboarding in
the exact Windows candidate. Rollback is the single runtime-boundary commit; it
does not alter existing configuration files.

## Open Questions

None for the focused runtime fix. Supported Windows packaging and identity-mode
choices remain program decisions in #2546.
