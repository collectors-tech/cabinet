# Cabinet infrastructure

`services/catalog.json` and each environment's `deployment-plan.json` are the
machine-readable sources of truth.

- `services/cabinet/compose.yaml` defines the reusable Cabinet service.
- `deployments/local/` defines local Docker Compose.
- `deployments/demo/` defines the self-hosted Coolify review environment.
- `deployments/production/` defines the explicitly approved Coolify release.

Operational instructions live in `docs/operations/deployments.md`.
