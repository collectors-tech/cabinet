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

test('backend owns the ZITADEL OIDC and application-session boundary', () => {
  const backend = read('internal/app/zitadel_auth.go')

  for (const fragment of [
    '/api/auth/zitadel/login',
    '/api/auth/zitadel/callback',
    '/api/auth/zitadel/session',
    '/api/auth/zitadel/refresh',
    '/api/auth/zitadel/logout',
    'code_challenge_method',
    'S256',
    'nonce',
    'jwks_uri',
    'cabinet_oidc_session',
    'HttpOnly: true',
    'SameSite: http.SameSiteLaxMode',
    'cabinet.user',
    'cabinet.admin',
  ]) {
    assert.ok(backend.includes(fragment), `missing backend contract: ${fragment}`)
  }

  assert.doesNotMatch(backend, /localStorage|sessionStorage/)
})

test('remote API requests fail closed behind the server session', () => {
  const app = read('internal/app/app.go')
  assert.match(app, /registerZitadelAuthRoutes/)
  assert.match(app, /requiresZitadelSession/)
  assert.match(app, /validateZitadelRequestSession/)
})

test('Cabinet sign-in remains branded and does not collect ZITADEL passwords', () => {
  const form = read(
    'ui.web/src/features/auth/sign-in/components/user-auth-form.tsx',
  )
  const authenticatedRoute = read('ui.web/src/routes/_authenticated/route.tsx')
  const signOutRoute = read('ui.web/src/routes/(auth)/sign-out.tsx')

  assert.match(form, /identityMode === 'zitadel'/)
  assert.match(form, /Continue securely/)
  assert.match(form, /\/api\/auth\/zitadel\/login/)
  assert.match(authenticatedRoute, /\/api\/auth\/zitadel\/session/)
  assert.match(signOutRoute, /\/api\/auth\/zitadel\/logout/)
})

test('all deployment environments enable isolated ZITADEL applications', () => {
  for (const environment of ['local', 'demo', 'production']) {
    const pattern =
      environment === 'local'
        ? `infra/deployments/${environment}/developer-machine/docker-compose/.env.example`
        : `infra/deployments/${environment}/selfhost-server/coolify/.env.example`
    const env = read(pattern)
    assert.match(env, /CABINET_AUTH_IDENTITY_MODE=zitadel/)
    assert.match(env, /CABINET_ZITADEL_LOGIN_V2_BASE_URL=/)
    assert.match(env, /CABINET_ZITADEL_CLIENT_ID=/)
    assert.match(env, /CABINET_ZITADEL_AUDIENCE=/)
    assert.match(env, /CABINET_ZITADEL_REQUIRED_ROLES=/)

    const identity = readJson(`infra/shared/identity/${environment}.json`)
    assert.equal(identity.loginExperience.implementation, 'zitadel-login-v2')
    assert.equal(identity.loginExperience.applicationSetting.useNewLoginUI, true)
    assert.equal(
      identity.loginExperience.applicationSetting.customBaseUrl,
      '${CABINET_ZITADEL_LOGIN_V2_BASE_URL}',
    )
    assert.equal(identity.loginExperience.trustedDomainRequired, true)
    assert.equal(
      identity.loginExperience.organisationBranding.projectPrivateLabeling,
      true,
    )
    assert.equal(
      identity.loginExperience.organisationBranding.disableWatermark,
      true,
    )
    assert.deepEqual(identity.acceptanceGates.readiness, [
      'issuer-discovery',
      'jwks',
      'login-v2',
      'callback',
      'provider-logout',
    ])
    assert.deepEqual(identity.acceptanceGates.deniedIdentities, [
      'wrong-issuer',
      'wrong-audience',
      'wrong-authorised-party',
      'expired-token',
      'missing-role',
      'unknown-signing-key',
    ])
  }
})
