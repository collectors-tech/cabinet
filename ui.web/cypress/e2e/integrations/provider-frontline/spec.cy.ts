describe('integrations/provider-frontline', () => {
  function signInToMarketWatch() {
    cy.visit('/sign-in?redirect=%2Fscanner%2F')
    cy.contains('button', 'Open local workspace').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
  }

  beforeEach(() => {
    cy.e2eReset()
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            key: 'frontlinehobbies',
            display_name: 'frontlinehobbies.com.au',
            provider_category: 'storefront/source matcher',
            market_watch_scope: 'frontlinehobbies',
            state: 'enabled',
            health: {
              status: 'ok',
              state: 'ready',
              label: 'Public Algolia storefront ready',
            },
            workflow_refs: ['market_watch.run'],
            capabilities: {
              search: true,
              pricing: true,
              stock_observation: true,
            },
          },
        ],
      },
    }).as('providerRegistry')
  })

  it('UC-AU-04 runs Frontline saved watches through the live provider route with provenance', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-frontline-live-1',
            name: 'Frontline AFX watch',
            keywords: ['AFX Falcon'],
            provider_scope: ['frontlinehobbies'],
            enabled: true,
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: { failures: [] },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', {
      statusCode: 500,
      body: { error: 'unexpected_generic_scanner_run_for_frontline' },
    }).as('genericScannerRun')
    cy.intercept('POST', '/api/providers/frontline/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-frontline-live-1')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'frontlinehobbies',
          query_set_id: 'qs-frontline-live-1',
          candidates: [
            {
              id: 'cand-frontline-live-1',
              query_set_id: 'qs-frontline-live-1',
              listing_id: 'frontline-afx-falcon',
              title: 'AFX Falcon Frontline Live Candidate',
              source: 'frontlinehobbies',
              price: 44.95,
              currency: 'AUD',
              url: 'https://frontlinehobbies.com.au/products/afx-falcon',
              seller: 'frontlinehobbies.com.au',
              stock_status: 'in_stock',
              handoff_state: 'wishlist_inventory_ready',
            },
          ],
          run_summary: {
            page_count: 1,
            observed_page_size: 1,
            candidates_total: 1,
          },
        },
      })
    }).as('frontlineProviderRun')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-frontline-live-1"]').click()
    cy.wait('@frontlineProviderRun')
    cy.get('@genericScannerRun.all').should('have.length', 0)
    cy.get('[data-testid="scanner-action-status"]').should(
      'contain',
      'frontline_run_started_qs-frontline-live-1'
    )
    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-frontline-live-1"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('td', 'frontlinehobbies').should('be.visible')
      cy.contains('td', 'AFX Falcon Frontline Live Candidate').should('be.visible')
      cy.contains('td', '44.95 AUD').should('be.visible')
      cy.contains(
        'td',
        'https://frontlinehobbies.com.au/products/afx-falcon'
      ).should('be.visible')
      cy.contains('td', 'in_stock').should('be.visible')
      cy.contains('td', 'wishlist_inventory_ready').should('be.visible')
    })
  })

  it('UC-AU-05 surfaces Frontline config-drift fallback warnings after cached recovery', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-frontline-drift-1',
            name: 'Frontline drift watch',
            keywords: ['AFX Mustang'],
            provider_scope: ['frontlinehobbies'],
            enabled: true,
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: { failures: [] },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/frontline/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-frontline-drift-1')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'frontlinehobbies',
          query_set_id: 'qs-frontline-drift-1',
          fallback_used: true,
          warning: 'Frontline discovery fallback used cached config.',
          candidates: [
            {
              id: 'cand-frontline-drift-1',
              query_set_id: 'qs-frontline-drift-1',
              listing_id: 'frontline-afx-mustang',
              title: 'AFX Mustang Frontline Recovered Candidate',
              source: 'frontlinehobbies',
              price: 49.95,
              currency: 'AUD',
              url: 'https://frontlinehobbies.com.au/products/afx-mustang',
              seller: 'frontlinehobbies.com.au',
              stock_status: 'in_stock',
              handoff_state: 'wishlist_inventory_ready',
            },
          ],
          run_summary: {
            page_count: 1,
            observed_page_size: 1,
            candidates_total: 1,
          },
        },
      })
    }).as('frontlineProviderRun')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-frontline-drift-1"]').click()
    cy.wait('@frontlineProviderRun')
    cy.get('[data-testid="scanner-action-status"]').should(
      'contain',
      'frontline_run_started_qs-frontline-drift-1'
    )
    cy.get('[data-testid="scanner-action-feedback"]').should(
      'contain',
      'Frontline discovery fallback used cached config.'
    )
  })
})
