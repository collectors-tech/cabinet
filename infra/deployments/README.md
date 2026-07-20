# Cabinet deployment environments

Cabinet has three isolated container deployments built from one reusable service
definition:

| Environment | Purpose | Orchestrator | Persistent data |
| --- | --- | --- | --- |
| Local Docker Compose | Developer and integration testing | Docker Compose | `cabinet-local-data` |
| Coolify demo | Max-facing review environment from validated `develop` work | Coolify | `cabinet-demo-data` |
| Coolify production | Explicitly approved self-hosted release | Coolify | `cabinet-production-data` |

The authoritative service catalogue is `infra/services/catalog.json`. Each
environment has one `infra/deployments/<environment>/deployment-plan.json`.
The local, demo and production workspaces never share a database, media tree,
backup directory, profile, instance name or Docker volume.

## Shared ZITADEL identity

Cabinet consumes the existing shared ZITADEL foundation as a pinned
`shared-reference`. It does not deploy ZITADEL from this repository and it
does not reuse another product's project, application, roles or credentials.

The per-environment contracts are under `infra/shared/identity/`. Local, demo
and production each require a distinct ZITADEL application, redirect URI,
post-logout URI, project audience and role assignment. Set
`CABINET_ZITADEL_AUDIENCE` to that environment's project ID so it matches the
requested project audience scope. The shared issuer readiness gate is
`/.well-known/openid-configuration`.

Application authentication is delivered under issue #1952. Set
`CABINET_AUTH_IDENTITY_MODE=zitadel` only after the corresponding environment
application, role grants, branded login and denied-identity tests pass. The
backend owns discovery, Authorization Code with PKCE, token validation,
refresh and logout; provider tokens never enter browser storage.

The user-facing route remains the custom Cabinet login at `/sign-in`. The
shared identity owner must configure ZITADEL Login V2 on the Cabinet custom
domain with Cabinet branding. For every application, enable **Use new login
UI**, set its custom base URL to `CABINET_ZITADEL_LOGIN_V2_BASE_URL`, add that
host to ZITADEL trusted domains, apply the organisation/project private-label
settings from the environment identity contract, and verify the URL before
cutover. Retain a break-glass `IAM_OWNER` identity so a bad Login V2 route can
be reverted. Cabinet does not collect the user's ZITADEL password and does not
deploy the shared login service. Provider tokens and secrets must not be stored
in browser local storage, committed environment examples, URLs, logs or
exported Cabinet data.

## Security and concurrency boundary

Cabinet is a local-first SQLite workspace, not a multi-tenant hosted service.
Every deployment runs a single replica. Coolify zero-downtime replacement must
be disabled so an old and a new container never overlap against the same
volume.

Local-device mode is not remote account authentication. A remote deployment
must not be exposed until ZITADEL readiness, exact redirects, role grants and
denied-identity checks pass. An approved access gate such as Tailscale or
Cloudflare Access remains recommended defence in depth. The
Compose services do not publish a host port on Coolify and must not mount the
Docker socket, repository checkout, host root or another environment's data.

## Local Docker Compose

From the repository root:

```powershell
Copy-Item infra/deployments/local/developer-machine/docker-compose/.env.example `
  infra/deployments/local/developer-machine/docker-compose/.env

docker compose `
  --env-file infra/deployments/local/developer-machine/docker-compose/.env `
  -f infra/deployments/local/developer-machine/docker-compose/compose.yaml `
  up --build -d
```

The local route is `http://127.0.0.1:17880`. Change
`CABINET_HOST_PORT` in the local environment file if that port is occupied.
The container always listens on internal port `17880`.

Verify the exact runtime:

```powershell
Invoke-WebRequest http://127.0.0.1:17880/healthz
Invoke-RestMethod http://127.0.0.1:17880/api/runtime
```

Stop without deleting the data volume:

```powershell
docker compose `
  --env-file infra/deployments/local/developer-machine/docker-compose/.env `
  -f infra/deployments/local/developer-machine/docker-compose/compose.yaml `
  down
```

Do not add `--volumes` unless the local workspace is intentionally being
destroyed.

## Build an immutable image

Demo and production consume an image digest, never a moving `latest`,
`develop` or version tag. Build from a clean, validated commit and pass the
same revision evidence into the runtime:

```powershell
$revision = git rev-parse HEAD
$buildDate = git show -s --format=%cI HEAD
$version = (Get-Content release/cabinet-beta-version.json -Raw | ConvertFrom-Json).version
$tag = "ghcr.io/collectors-tech/cabinet:sha-$revision"

docker build `
  --build-arg "CABINET_BUILD_VERSION=$version" `
  --build-arg "CABINET_BUILD_REVISION=$revision" `
  --build-arg "CABINET_BUILD_DATE=$buildDate" `
  --tag $tag .

docker push $tag
docker buildx imagetools inspect $tag
```

Copy the resulting
`ghcr.io/collectors-tech/cabinet@sha256:<64-character-digest>` reference into
the applicable Coolify `CABINET_IMAGE` variable and deployment
`image-lock.json`. The image digest, source revision and `/api/runtime`
response must agree.

## Coolify demo

Create one Docker Compose resource:

- Repository: `collectors-tech/cabinet`
- Branch: `develop`
- Compose path:
  `infra/deployments/demo/selfhost-server/coolify/compose.yaml`
- Container port: `17880`
- Health path: `/healthz`
- Persistent volume: `cabinet-demo-data:/data`
- Replicas: one
- Zero-downtime deployment: disabled

Copy the keys from the adjacent `.env.example` into Coolify and replace every
placeholder. Set `CABINET_PUBLIC_ORIGIN` to the exact HTTPS URL and
`CABINET_WEBAUTHN_RP_ID` to its hostname. Configure the domain to proxy to
port `17880`, provision the matching Cabinet Demo ZITADEL application and
roles, then apply the access gate before allowing remote traffic.

Demo promotion is from a green `develop` revision only. It does not authorise
merging `develop` into `main`.

## Coolify production

Create a separate Docker Compose resource:

- Compose path:
  `infra/deployments/production/selfhost-server/coolify/compose.yaml`
- Persistent volume: `cabinet-production-data:/data`
- Image: explicitly approved immutable digest
- Replicas: one
- Zero-downtime deployment: disabled

Production must use its own domain, access boundary, environment values and
volume. It must not reuse the demo resource or data. Deployment requires Max's
explicit release approval; creating this resource does not grant approval to
merge `develop` into `main`.

For each remote environment, verify that the issuer discovery document, JWKS,
`CABINET_ZITADEL_LOGIN_V2_BASE_URL`, Cabinet-branded login, callback and
provider logout are reachable before deployment. A user with the wrong
audience, authorised party, issuer, expired token or missing Cabinet role must
be denied.

## Backup and restore

Before every demo or production upgrade:

1. Call Cabinet's `POST /api/backup/run`.
2. Confirm the new backup through `GET /api/backup/list`.
3. Copy the resulting backup and required media outside the active Docker
   volume using the approved host backup process.
4. Record the current image digest, source revision and volume name.

Do not copy a live `cabinet.db` file as the only backup. A backup stored only
inside `/data` does not protect against loss of the Docker host or volume.

Restore into an isolated replacement volume first, validate record/media
relationships, then switch the deployment. Do not test a restore by overwriting
the active demo or production workspace.

## Upgrade verification

After every deployment:

1. Confirm `/healthz` returns `200 ok`.
2. Confirm `/api/runtime` reports the expected app version, build date,
   internal port and `/data` data directory.
3. Restart the resource and confirm the same records and media remain.
4. Confirm demo and production volumes have not changed or crossed.
5. Exercise one backup and list operation.
6. Confirm remote access is rejected before the configured access gate.

## Rollback

1. Stop the current container without removing its volume.
2. Set `CABINET_IMAGE` back to the last validated image digest.
3. Redeploy with zero-downtime replacement still disabled.
4. Verify `/healthz`, `/api/runtime` and existing data.
5. If a migration prevents binary rollback, restore the pre-upgrade backup into
   a new isolated volume and retain the failed volume for investigation.

Never use `docker compose down --volumes` or delete a Coolify persistent
volume as part of a normal rollback.
