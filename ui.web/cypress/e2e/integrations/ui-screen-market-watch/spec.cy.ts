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

  it('UI-SCREEN-MARKET-WATCH-006 runs Bonza AFX query and surfaces aggregated run summary', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-bonza',
            name: 'AFX',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
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
    cy.intercept('POST', '/api/providers/bonza/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-bonza')
      req.reply({
        statusCode: 200,
        body: {
          query_set_id: 'qs-mw-bonza',
          page_count: 2,
          observed_page_size: 2,
          items_per_page_used: 36,
          candidates: [
            { listing_id: 'bonza-1', title: 'AFX Camaro', source: 'bonzaslotcars' },
            { listing_id: 'bonza-2', title: 'AFX Mustang', source: 'bonzaslotcars' },
            { listing_id: 'bonza-3', title: 'AFX Mega G+', source: 'bonzaslotcars' },
          ],
          run_summary: {
            page_count: 2,
            observed_page_size: 2,
            candidates_total: 3,
          },
        },
      })
    }).as('runBonzaQuery')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-bonza"]').click()
    cy.wait('@runBonzaQuery')
    cy.get('[data-testid="scanner-run-summary-qs-mw-bonza"]')
      .should('contain', 'Pages: 2')
      .and('contain', 'Candidates: 3')
      .and('contain', 'Observed page size: 2')
  })

  it('UI-SCREEN-MARKET-WATCH-005 renders query table view with saved-query columns for rapid inspection', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-table-1',
            name: 'Bonza AFX Watch',
            keywords: ['AFX', 'Mega G+'],
            provider_scope: ['bonzaslotcars'],
          },
          {
            id: 'qs-mw-table-2',
            name: 'eBay HO Scan',
            keywords: ['HO slot'],
            provider_scope: ['ebay'],
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

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-query-table"]').should('be.visible')
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('th', 'Query Name').should('be.visible')
      cy.contains('th', 'Provider Scope').should('be.visible')
      cy.contains('th', 'Last Run Status').should('be.visible')
      cy.contains('th', 'Last Run Time').should('be.visible')
      cy.contains('th', 'Latest Output Summary').should('be.visible')
      cy.contains('td', 'Bonza AFX Watch').should('be.visible')
      cy.contains('td', 'bonzaslotcars').should('be.visible')
    })
  })

  it('UI-SCREEN-MARKET-WATCH-005 opens deterministic output details from query-table row action', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-detail-1',
            name: 'Bonza AFX Watch',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
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
    cy.intercept('POST', '/api/providers/bonza/run', {
      statusCode: 200,
      body: {
        query_set_id: 'qs-mw-detail-1',
        page_count: 2,
        observed_page_size: 2,
        items_per_page_used: 36,
        candidates: [
          { listing_id: 'bonza-1', title: 'AFX Camaro', source: 'bonzaslotcars' },
          { listing_id: 'bonza-2', title: 'AFX Mustang', source: 'bonzaslotcars' },
        ],
        run_summary: {
          page_count: 2,
          observed_page_size: 2,
          candidates_total: 2,
        },
      },
    }).as('runBonzaQuery')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-mw-detail-1"]').click()
    cy.wait('@runBonzaQuery')

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-detail-1"]').click()
    cy.get('[data-testid="market-watch-output-detail"]').within(() => {
      cy.contains('Bonza AFX Watch').should('be.visible')
      cy.contains('Provider Scope').should('be.visible')
      cy.contains('bonzaslotcars').should('be.visible')
      cy.contains('Pages scanned').should('be.visible')
      cy.contains('2').should('be.visible')
      cy.contains('Candidates').should('be.visible')
    })
  })
})
