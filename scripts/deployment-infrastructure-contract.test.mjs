import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'

const repoRoot = resolve(import.meta.dirname, '..')

function read(relativePath) {
  return readFileSync(resolve(repoRoot, relativePath), 'utf8')
}

function readJson(relativePath) {
  return JSON.parse(read(relativePath))
}

const environments = [
  {
    id: 'local',
    deploymentId: 'cabinet-local',
    target: 'developer-machine',
    orchestrator: 'docker-compose',
    compose:
      'infra/deployments/local/developer-machine/docker-compose/compose.yaml',
    envExample:
      'infra/deployments/local/developer-machine/docker-compose/.env.example',
    volume: 'cabinet-local-data',
    profile: 'local-compose',
  },
  {
    id: 'demo',
    deploymentId: 'cabinet-demo',
    target: 'selfhost-server',
    orchestrator: 'coolify',
    compose:
      'infra/deployments/demo/selfhost-server/coolify/compose.yaml',
    envExample:
      'infra/deployments/demo/selfhost-server/coolify/.env.example',
    volume: 'cabinet-demo-data',
    profile: 'demo',
  },
  {
    id: 'production',
    deploymentId: 'cabinet-production',
    target: 'selfhost-server',
    orchestrator: 'coolify',
    compose:
      'infra/deployments/production/selfhost-server/coolify/compose.yaml',
    envExample:
      'infra/deployments/production/selfhost-server/coolify/.env.example',
    volume: 'cabinet-production-data',
    profile: 'production',
  },
]

test('catalogue defines one reusable single-replica Cabinet service', () => {
  const catalogue = readJson('infra/services/catalog.json')

  assert.equal(catalogue.schemaVersion, 1)
  assert.equal(catalogue.services.length, 1)

  const service = catalogue.services[0]
  assert.equal(service.id, 'cabinet')
  assert.equal(
    service.canonicalCompose,
    'infra/services/cabinet/compose.yaml',
  )
  assert.equal(service.constraints.replicas, 1)
  assert.equal(service.persistence[0].mountPath, '/data')
  assert.equal(service.health.liveness.path, '/healthz')
  assert.equal(service.health.readiness.path, '/api/runtime')
  assert.deepEqual(service.dependencies, [])
  assert.deepEqual(service.oneShots, [])

  assert.equal(catalogue.sharedReferences.length, 1)
  const identity = catalogue.sharedReferences[0]
  assert.equal(identity.id, 'shared-zitadel')
  assert.equal(identity.kind, 'shared-reference')
  assert.equal(identity.service, 'zitadel')
  assert.equal(identity.capability, 'oidc')
  assert.equal(identity.owner, 'platform-infrastructure')
  assert.match(identity.contractRevision, /973bda4acdd3b4c81356dc2145ac34a6397222d0/)
})

test('local, demo and production plans resolve one deployment in one layer', () => {
  const deploymentIds = new Set()

  for (const environment of environments) {
    const plan = readJson(
      `infra/deployments/${environment.id}/deployment-plan.json`,
    )

    assert.equal(plan.schemaVersion, 1)
    assert.equal(plan.environment, environment.id)
    assert.equal(plan.deployments.length, 1)
    assert.equal(plan.layers.length, 1)

    const deployment = plan.deployments[0]
    assert.equal(deployment.id, environment.deploymentId)
    assert.equal(deployment.target, environment.target)
    assert.equal(deployment.orchestrator, environment.orchestrator)
    assert.deepEqual(deployment.source, { type: 'service', id: 'cabinet' })
    assert.equal(deployment.compose, environment.compose)
    assert.equal(deployment.replicas, 1)
    assert.deepEqual(deployment.dependencies, [])
    assert.equal(deployment.zeroDowntime, false)
    assert.equal(plan.consumes.length, 1)
    assert.equal(plan.consumes[0].service, 'zitadel')
    assert.equal(plan.consumes[0].mode, 'shared-reference')
    assert.equal(
      plan.consumes[0].projectConfiguration,
      `infra/shared/identity/${environment.id}.json`,
    )

    assert.deepEqual(plan.layers[0].deployments, [environment.deploymentId])
    assert.equal(deploymentIds.has(deployment.id), false)
    deploymentIds.add(deployment.id)
  }
})

test('Compose deployments share the canonical service and isolate runtime state', () => {
  const canonical = read('infra/services/cabinet/compose.yaml')
  for (const fragment of [
    'CABINET_BIND_MODE: lan',
    'CABINET_HOST: 0.0.0.0',
    'CABINET_PORT: "17880"',
    'CABINET_DATA_DIR: /data',
    'CABINET_DB_PATH: /data/cabinet.db',
    'CABINET_ALLOW_PARALLEL: "false"',
    'CABINET_E2E_MODE: "false"',
    'no-new-privileges:true',
    'cap_drop:',
    '- ALL',
    'stop_grace_period: 30s',
    'http://127.0.0.1:17880/healthz',
  ]) {
    assert.match(canonical, new RegExp(fragment.replaceAll('/', '\\/')))
  }

  assert.doesNotMatch(canonical, /--allow-parallel|--seed-sample-data/)
  assert.doesNotMatch(canonical, /docker\.sock|\/[\s]*:\/[\s]*/)

  for (const environment of environments) {
    const compose = read(environment.compose)
    const envExample = read(environment.envExample)

    assert.match(compose, /extends:/)
    assert.match(compose, /services\/cabinet\/compose\.yaml/)
    assert.match(compose, new RegExp(environment.volume))
    assert.match(envExample, new RegExp(`CABINET_PROFILE=${environment.profile}`))
    assert.match(
      envExample,
      new RegExp(`CABINET_INSTANCE_NAME=${environment.deploymentId}`),
    )

    if (environment.id === 'local') {
      assert.match(compose, /build:/)
      assert.match(compose, /127\.0\.0\.1:\$\{CABINET_HOST_PORT:-17880\}:17880/)
    } else {
      assert.doesNotMatch(compose, /ports:/)
      assert.match(envExample, /CABINET_IMAGE=ghcr\.io\/collectors-tech\/cabinet@sha256:/)
      assert.match(envExample, /CABINET_ACCESS_GATE=/)
    }
  }
})

test('container defaults are deployment-neutral and accept build metadata', () => {
  const dockerfile = read('Dockerfile')

  for (const fragment of [
    'ARG CABINET_BUILD_VERSION',
    'ARG CABINET_BUILD_REVISION',
    'ARG CABINET_BUILD_DATE',
    'internal/app.buildVersion=${CABINET_BUILD_VERSION}',
    'internal/app.buildRevision=${CABINET_BUILD_REVISION}',
    'internal/app.buildDate=${CABINET_BUILD_DATE}',
    '--profile", "cabinet',
    '--instance-name", "cabinet-container',
  ]) {
    assert.ok(
      dockerfile.includes(fragment),
      `Dockerfile missing required fragment: ${fragment}`,
    )
  }

  assert.doesNotMatch(dockerfile, /e2e-cypress|cypress-container|--allow-parallel/)
})

test('operator documentation covers all environment safety boundaries', () => {
  const docs = read('infra/deployments/README.md')

  for (const fragment of [
    'Local Docker Compose',
    'Coolify demo',
    'Coolify production',
    'single replica',
    'zero-downtime',
    'access gate',
    'backup',
    'restore',
    'rollback',
    '/api/runtime',
    'image digest',
    'shared zitadel',
    'custom cabinet login',
  ]) {
    assert.match(docs.toLowerCase(), new RegExp(fragment.toLowerCase()))
  }
})

test('Docker-capable gates render every Compose deployment', () => {
  const validator = read('scripts/validate-compose-deployments.sh')
  assert.match(validator, /docker compose/)
  for (const environment of environments) {
    assert.ok(
      validator.includes(environment.compose),
      `Compose validator missing ${environment.compose}`,
    )
    assert.ok(
      validator.includes(environment.envExample),
      `Compose validator missing ${environment.envExample}`,
    )
  }

  for (const workflow of [
    '.github/workflows/develop-quality-gate.yml',
    '.github/workflows/main-gate.yml',
  ]) {
    const content = read(workflow)
    assert.match(content, /Validate Docker Compose deployments/)
    assert.match(content, /bash scripts\/validate-compose-deployments\.sh/)
  }
})
