describe('settings/integrations MCP transport', () => {
  function signInToIntegrations() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fintegrations')
    cy.contains('button', 'Open local workspace').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/integrations\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SETTINGS-INTEGRATIONS-MCP-001 configures loopback transport without leaking stored credential on status refresh', () => {
    const profileId = 'profile-mcp-ui'
    const generatedCredential = 'cabinet-mcp-secret-visible-once'
    let statusAttempt = 0

    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: profileId },
    }).as('activeProfile')
    cy.intercept('GET', `/api/profiles/${profileId}/settings`, {
      statusCode: 200,
      body: { settings: {} },
    }).as('profileSettings')
    cy.intercept('GET', `/api/profiles/${profileId}/mcp-http-status`, (req) => {
      statusAttempt += 1
      req.reply({
        statusCode: 200,
        body:
          statusAttempt <= 2
            ? {
                enabled: false,
                state: 'disabled',
                listen_addr: '127.0.0.1:17890',
                credential_configured: false,
                guidance: 'HTTP transport is disabled.',
                recovery_action: 'Enable the transport for local clients.',
              }
            : {
                enabled: true,
                state: 'ready',
                listen_addr: '127.0.0.1:17891',
                credential_configured: true,
                guidance: 'Loopback HTTP transport is ready.',
                recovery_action: '',
              },
      })
    }).as('mcpStatus')
    cy.intercept('PUT', `/api/profiles/${profileId}/mcp-http-config`, (req) => {
      expect(req.body).to.deep.equal({
        enabled: true,
        listen_addr: '127.0.0.1:17891',
      })
      req.reply({
        statusCode: 200,
        body: {
          enabled: true,
          state: 'misconfigured',
          listen_addr: '127.0.0.1:17891',
          credential_configured: false,
          guidance: 'Generate a credential before connecting clients.',
          recovery_action: 'Generate credential.',
        },
      })
    }).as('mcpConfig')
    cy.intercept(
      'POST',
      `/api/profiles/${profileId}/mcp-http-credential`,
      {
        statusCode: 200,
        body: {
          credential: generatedCredential,
          credential_configured: true,
          configuration_guidance:
            'Copy this credential now; it will not be shown again.',
        },
      }
    ).as('mcpCredential')

    signInToIntegrations()
    cy.wait('@activeProfile')
    cy.wait('@profileSettings')
    cy.wait('@mcpStatus')

    cy.get('[data-testid="settings-integrations-mcp-card"]').should(
      'contain',
      'Cabinet MCP'
    )
    cy.get('[data-testid="settings-integrations-mcp-state"]').should(
      'contain',
      'Disabled'
    )
    cy.get('[data-testid="settings-integrations-mcp-profile"]').should(
      'contain',
      profileId
    )
    cy.get('[data-testid="settings-integrations-mcp-credential-state"]').should(
      'contain',
      'Missing'
    )

    cy.get('[data-testid="settings-integrations-mcp-listen"]')
      .clear()
      .type('127.0.0.1:17891')
    cy.get('[data-testid="settings-integrations-mcp-enabled"]').click()
    cy.wait('@mcpConfig')
    cy.get('[data-testid="settings-integrations-mcp-state"]').should(
      'contain',
      'Misconfigured'
    )

    cy.get('[data-testid="settings-integrations-mcp-generate-credential"]').click()
    cy.wait('@mcpCredential')
    cy.wait('@mcpStatus')
    cy.get('[data-testid="settings-integrations-mcp-generated-credential"]').should(
      'contain',
      generatedCredential
    )
    cy.get('[data-testid="settings-integrations-mcp-credential-state"]').should(
      'contain',
      'Configured'
    )

    cy.get('[data-testid="settings-integrations-mcp-refresh"]').click()
    cy.wait('@mcpStatus')
    cy.get('[data-testid="settings-integrations-mcp-generated-credential"]').should(
      'contain',
      'Credential appears only immediately after setup.'
    )
    cy.get('[data-testid="settings-integrations-mcp-card"]').should(
      'not.contain',
      generatedCredential
    )
    cy.location('pathname').should('match', /^\/settings\/integrations\/?$/)
  })
})
