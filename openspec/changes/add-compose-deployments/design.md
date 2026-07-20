## Context

Cabinet stores its SQLite database, media, attachments, backups and runtime
state under one data directory. A self-hosted deployment must preserve that
local-first boundary and must never run overlapping replicas against one data
volume.

## Goals / Non-Goals

**Goals:**

- Reuse one canonical service definition in all environments.
- Isolate local, demo and production identities and data.
- Make Coolify deployment inputs explicit and machine-readable.
- Require deployment revision and immutable image evidence.

**Non-Goals:**

- Multi-tenant hosting.
- Introducing separate database, cache, object-storage or reverse-proxy
  services.
- Automating a merge from `develop` into `main`.

## Decisions

1. Model demo as a real environment rather than calling it production.
   - Demo runs a production build but has review/reset and access policies that
     differ from an approved production workspace.

2. Use one canonical Compose service with environment deployments extending it.
   - This keeps health, runtime command and security defaults in one place
     while volumes, ports and build/image policy remain environment-specific.

3. Enforce one replica and disable overlapping replacement.
   - SQLite and the shared data root require one active writer per deployment.

4. Bind local Compose only to loopback and expose no host port in Coolify
   Compose.
   - Coolify owns reverse proxy and TLS routing; remote environments also
     require an approved access gate.

5. Pin demo and production to image digests.
   - Tags aid discovery but do not provide immutable deployment evidence.

6. Consume shared ZITADEL rather than adding it to Cabinet Compose.
   - Cabinet owns separate local, demo and production application
     configuration while platform infrastructure owns the shared service.

## Risks / Trade-offs

- Some hosted Compose implementations may not support `extends.file`.
  - Mitigation: validate against the Coolify Docker Compose version before live
    promotion; retain the canonical service as the source for any generated
    flattened file if Coolify requires it.

- Application backups live beneath the same data root.
  - Mitigation: require an off-volume copy and isolated restore proof.

- Local-device mode does not authenticate remote users.
  - Mitigation: require Tailscale, Cloudflare Access or another approved gate
    before exposing demo or production.

## Migration Plan

1. Land catalogue, plans, Compose inputs and contracts.
2. Build and publish an image from a validated commit with revision metadata.
3. Configure the isolated Coolify demo resource and capture live evidence.
4. Validate backup, restart, upgrade and rollback.
5. Create production only from an explicitly approved release digest.

Rollback:

- Stop the active container, restore the previous image digest and preserve the
  existing volume. Restore the pre-upgrade backup into a new volume if schema
  compatibility prevents binary rollback.
