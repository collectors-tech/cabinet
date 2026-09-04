## Why

Cabinet has a reusable Linux container image but no authoritative Docker
Compose or Coolify deployment model. Local container testing, self-hosted demo
review and approved production deployment therefore cannot be reproduced or
validated as isolated environments.

## What Changes

- Define the canonical Cabinet service and its health, persistence and
  single-replica constraints.
- Add local, demo and production deployment plans.
- Add a local build-based Compose deployment and digest-pinned Coolify
  deployments.
- Remove E2E identity and parallel-runtime flags from production image defaults.
- Add deployment metadata, isolation, access, backup and rollback contracts.
- Consume the shared ZITADEL foundation through a pinned per-environment
  reference without deploying a duplicate identity service.
- Extend the existing develop quality gate with the infrastructure contract.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `runtime-core`: Cabinet gains reproducible local, demo and production
  container deployment contracts.

## Impact

- Affected code: `Dockerfile`, `infra/`, deployment contract validation.
- Affected operations: local Docker Compose and Coolify demo/production
  resources.
- Affected documentation: README and deployment runbook.
- Related issue: `#1951`.
