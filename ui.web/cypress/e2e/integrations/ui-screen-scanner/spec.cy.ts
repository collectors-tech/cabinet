describe('integrations/ui-screen-scanner', () => {
  function signInToScanner() {
    cy.visit('/sign-in?redirect=%2Fscanner%2F')
    cy.get('input[name="email"]').clear().type('e2e-scanner@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SCANNER-005 uses Market Watch naming with scanner route compatibility', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: { query_sets: [] },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.location('pathname').should('match', /^\/scanner\/?$/)
    cy.get('[data-testid="sidebar-nav-link-market-watch"]').should('contain', 'Market Watch')
    cy.get('[data-testid="sidebar-nav-link-market-watch"]').should('not.contain', 'Scanner')
    cy.contains('h1', 'Market Watch').should('be.visible')
  })

  it('UI-SCREEN-SCANNER-001 supports query set create/load and run controls', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [{ id: 'qs-1', name: 'AFX Core', keywords: ['afx', 'mega g'] }],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/query-sets', {
      statusCode: 201,
      body: { id: 'qs-2', name: 'New Set', keywords: ['camaro'] },
    }).as('createQuerySet')
    cy.intercept('POST', '/api/scanner/run', {
      statusCode: 200,
      body: { run_id: 'run-1', status: 'ok' },
    }).as('runNow')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.contains('[data-testid="scanner-query-list"]', 'AFX Core').should('be.visible')

    cy.get('[data-testid="scanner-new-query-name"]').type('New Set')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('camaro')
    cy.get('[data-testid="scanner-create-query"]').click()
    cy.wait('@createQuerySet')
    cy.get('[data-testid="scanner-action-status"]').should('contain', 'query_set_created')

    cy.get('[data-testid="scanner-run-qs-1"]').click()
    cy.wait('@runNow')
    cy.get('[data-testid="scanner-action-status"]').should('contain', 'run_started_qs-1')
  })

  it('UI-SCREEN-SCANNER-002 exposes provider health and failure retry', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [{ id: 'qs-9', name: 'Retry Set', keywords: ['retry'] }],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: {
        failures: [{ id: 'failure-1', query_set_id: 'qs-9', provider: 'ebay', message: 'timeout' }],
      },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'degraded' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/failures/retry', {
      statusCode: 200,
      body: { status: 'retry_requested', query_set_id: 'qs-9' },
    }).as('retryFailure')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-provider-health"]').should('contain', 'degraded')
    cy.get('[data-testid="scanner-retry-qs-9"]').click()
    cy.wait('@retryFailure')
    cy.get('[data-testid="scanner-action-status"]').should('contain', 'retry_requested_qs-9')
  })

  it('UI-SCREEN-SCANNER-003 shows deterministic empty and error states', () => {
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as(
      'emptyQuerySets'
    )
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'emptyFailures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToScanner()
    cy.wait(['@emptyQuerySets', '@emptyFailures', '@providerHealth'])
    cy.get('[data-testid="scanner-empty-state"]').should('be.visible')

    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 500,
      body: { error: 'failed_to_list_query_sets' },
    }).as('querySetsError')
    cy.reload()
    cy.wait('@querySetsError')
    cy.get('[data-testid="scanner-error-state"]').should('be.visible')
  })

  it('UI-SCREEN-SCANNER-004 maps run failures to actionable guidance', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [{ id: 'qs-bad', name: 'Broken Query', keywords: ['bad'] }],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'degraded' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', {
      statusCode: 400,
      body: { error: 'query_validation_failed' },
    }).as('runNowFailed')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-bad"]').click()
    cy.wait('@runNowFailed')

    cy.get('[data-testid="scanner-action-feedback"]')
      .should('be.visible')
      .and('contain', 'Run failed due to query validation.')
      .and('contain', 'Check query keywords and exclusions.')
      .and('contain', 'Review provider health and credentials before retrying.')
    cy.get('[data-testid="scanner-action-feedback"]').should('not.contain', 'run_failed_400')
    cy.get('[data-testid="scanner-action-diagnostics"]')
      .should('be.visible')
      .and('contain', 'query_validation_failed')
  })
})
