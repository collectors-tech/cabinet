describe('Browser Companion pairing security', () => {
  const pairingRequest = {
    request_id: 'pairing-1',
    pairing_code: '472981',
    device_id: 'chrome-windows-1',
    device_name: 'Chrome on Windows',
    origin: 'chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
    protocol_version: '1',
    capabilities: ['modules:read', 'captures:submit'],
    status: 'pending',
    expires_at: '2026-08-06T12:10:00Z',
    created_at: '2026-08-06T12:05:00Z',
  }

  const session = {
    id: 'session-1',
    cabinet_instance_id: 'cabinet-instance-1',
    profile_id: 'profile-e2e-001',
    device_id: 'edge-windows-1',
    device_name: 'Edge on Windows',
    origin: 'chrome-extension://bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb',
    protocol_version: '1',
    capabilities: ['modules:read'],
    created_at: '2026-08-01T10:00:00Z',
    expires_at: '2026-09-01T10:00:00Z',
    last_used_at: '2026-08-06T11:00:00Z',
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: { providers: [] },
    }).as('registry')
    cy.intercept('GET', '/api/profiles/profile-e2e-001/settings', {
      statusCode: 200,
      body: { settings: {} },
    }).as('settings')
    cy.intercept('GET', '/api/companion/pairing/requests', {
      statusCode: 200,
      body: { requests: [pairingRequest] },
    }).as('pairingRequests')
    cy.intercept('GET', '/api/companion/sessions', {
      statusCode: 200,
      body: { sessions: [session] },
    }).as('companionSessions')

    cy.visit('/sign-in?redirect=%2Fintegrations%2F')
    cy.contains('button', 'Open local workspace').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/integrations\/?$/
    )
    cy.wait([
      '@activeProfile',
      '@registry',
      '@settings',
      '@pairingRequests',
      '@companionSessions',
    ])
  })

  it('requires explicit in-Cabinet approval and keeps credentials redacted', () => {
    cy.intercept('POST', '/api/companion/pairing/approvals', {
      statusCode: 200,
      body: { ...pairingRequest, status: 'approved' },
    }).as('approvePairing')

    cy.get('[data-testid="browser-companion-security"]')
      .should('contain', 'Approve only an extension and pairing code you recognise')
      .and('contain', 'Chrome on Windows')
      .and('contain', 'Code 472981')
      .and('not.contain', 'credential')
      .and('not.contain', 'exchange_secret')

    cy.get('[data-testid="browser-companion-pairing-pairing-1"]')
      .contains('button', 'Approve')
      .click()
    cy.wait('@approvePairing').its('request.body').should('deep.equal', {
      request_id: 'pairing-1',
      profile_id: 'profile-e2e-001',
      capabilities: ['modules:read', 'captures:submit'],
    })
    cy.contains('Approved Browser Companion Chrome on Windows.').should(
      'be.visible'
    )
  })

  it('revokes one session or every profile session without exposing a secret', () => {
    cy.intercept('DELETE', '/api/companion/sessions?id=session-1', {
      statusCode: 200,
      body: { revoked_count: 1 },
    }).as('revokeSession')
    cy.intercept('DELETE', '/api/companion/sessions?all=true', {
      statusCode: 200,
      body: { revoked_count: 1 },
    }).as('revokeAll')

    cy.get('[data-testid="browser-companion-session-session-1"]')
      .should('contain', 'Edge on Windows')
      .and('not.contain', 'Bearer')
      .contains('button', 'Revoke')
      .click()
    cy.wait('@revokeSession')
    cy.contains('Revoked the selected Browser Companion session.').should(
      'be.visible'
    )

    cy.contains('button', 'Revoke all').click()
    cy.wait('@revokeAll')
  })
})
