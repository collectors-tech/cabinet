#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

deployments=(
  "infra/deployments/local/developer-machine/docker-compose/compose.yaml|infra/deployments/local/developer-machine/docker-compose/.env.example"
  "infra/deployments/demo/selfhost-server/coolify/compose.yaml|infra/deployments/demo/selfhost-server/coolify/.env.example"
  "infra/deployments/production/selfhost-server/coolify/compose.yaml|infra/deployments/production/selfhost-server/coolify/.env.example"
)

for deployment in "${deployments[@]}"; do
  IFS='|' read -r compose_file env_file <<<"$deployment"
  echo "Validating $compose_file"
  docker compose \
    --env-file "$env_file" \
    --file "$compose_file" \
    config --quiet
done
