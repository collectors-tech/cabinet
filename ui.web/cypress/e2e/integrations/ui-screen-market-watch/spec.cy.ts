describe('integrations/ui-screen-market-watch', () => {
  function signInToMarketWatch() {
    cy.visit('/sign-in?redirect=%2Fscanner%2F')
    cy.get('input[name="email"]').clear().type('e2e-market-watch@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-MARKET-WATCH-001 creates provider-scoped query sets from selector controls', () => {
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as(
      'querySets'
    )
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/query-sets', (req) => {
      expect(req.body.provider_scope).to.deep.equal(['amazon'])
      req.reply({
        statusCode: 201,
        body: {
          id: 'qs-mw-1',
          name: 'Amazon Scope',
          keywords: ['slot'],
          provider_scope: ['amazon'],
        },
      })
    }).as('createScopedQuery')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-provider-mode"]').should('have.value', 'single')
    cy.get('[data-testid="market-watch-provider-single"]').select('amazon')
    cy.get('[data-testid="scanner-new-query-name"]').type('Amazon Scope')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('slot')
    cy.get('[data-testid="scanner-create-query"]').click()
    cy.wait('@createScopedQuery')
    cy.get('[data-testid="scanner-query-providers-qs-mw-1"]').should('contain', 'amazon')
  })

  it('UI-SCREEN-MARKET-WATCH-002 sends provider scope in run payload and shows provider-attributed results', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-2',
            name: 'Dual Scope',
            keywords: ['ho slot'],
            provider_scope: ['ebay', 'amazon'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-2')
      expect(req.body.provider_scope).to.deep.equal(['ebay', 'amazon'])
      req.reply({ statusCode: 200, body: { run_id: 'run-mw-2', status: 'ok' } })
    }).as('runScopedQuery')
    cy.intercept('GET', '/api/scanner/candidates?query_set_id=qs-mw-2', {
      statusCode: 200,
      body: {
        candidates: [
          {
            id: 'cand-1',
            query_set_id: 'qs-mw-2',
            listing_id: 'L-1',
            title: 'AFX Camaro',
            source: 'amazon',
          },
        ],
      },
    }).as('listCandidates')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-2"]').click()
    cy.wait('@runScopedQuery')
    cy.wait('@listCandidates')
    cy.get('[data-testid="scanner-candidates-qs-mw-2"]').should('contain', 'amazon')
  })

  it('UI-SCREEN-MARKET-WATCH-003 blocks create when provider scope is empty in multi-provider mode', () => {
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as(
      'querySets'
    )
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-provider-mode"]').select('multi')
    cy.get('[data-testid="market-watch-provider-checkbox-ebay"]').uncheck({ force: true })
    cy.get('[data-testid="market-watch-provider-checkbox-amazon"]').uncheck({ force: true })
    cy.get('[data-testid="scanner-new-query-name"]').type('No Provider Query')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('slot')
    cy.get('[data-testid="scanner-create-query"]').click()

    cy.get('[data-testid="market-watch-provider-validation"]')
      .should('be.visible')
      .and('contain', 'Select at least one provider')
  })
})
