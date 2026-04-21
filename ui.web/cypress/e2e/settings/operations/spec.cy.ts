describe('settings/operations', () => {
  function signInToOperations() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Foperations')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/operations\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-001 renders runtime metadata and recovery visibility', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-test123',
        build_date: '2026-04-22',
        bind_mode: 'lan',
        runtime_host: '192.168.1.53',
        runtime_port: 17882,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: true,
      },
    }).as('runtimeRecovery')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'rev-test123'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      '192.168.1.53:17882'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'lan'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'stable'
    )
    cy.get('[data-testid="settings-operations-recovery-card"]').should(
      'contain',
      'Recovery required'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-002 retries runtime visibility without route reload', () => {
    let runtimeAttempt = 0
    let recoveryAttempt = 0

    cy.intercept('GET', '/api/runtime', (req) => {
      runtimeAttempt += 1
      if (runtimeAttempt === 1) {
        req.reply(503, { error: 'runtime_unavailable' })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          app_version: 'rev-recovered',
          build_date: '2026-04-22',
          bind_mode: 'loopback',
          runtime_host: '127.0.0.1',
          runtime_port: 17880,
          update_channel: 'stable',
          update_public_key_configured: false,
        },
      })
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', (req) => {
      recoveryAttempt += 1
      if (recoveryAttempt === 1) {
        req.reply(503, { error: 'runtime_recovery_unavailable' })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          recovery_required: false,
        },
      })
    }).as('runtimeRecovery')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.get('[data-testid="settings-operations-runtime-error"]').should(
      'contain',
      'Runtime information is unavailable right now.'
    )

    cy.get('[data-testid="settings-operations-retry"]').click()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-runtime-error"]').should('not.exist')
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'rev-recovered'
    )
    cy.get('[data-testid="settings-operations-recovery-card"]').should(
      'contain',
      'No recovery required'
    )
  })
})
