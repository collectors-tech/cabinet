describe('integrations/provider-hobbytech', () => {
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
            key: 'hobbytechtoys',
            display_name: 'hobbytechtoys.com.au',
            provider_category: 'storefront/source matcher',
            market_watch_scope: 'hobbytechtoys',
            state: 'enabled',
            health: {
              status: 'ok',
              state: 'ready',
              label: 'Public storefront ready',
            },
            workflow_refs: ['market_watch.run', 'hobbytech.parts_finder'],
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

  it('UC-AU-06 runs Hobbytech saved watches through the live provider route with provenance', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-hobbytech-live-1',
            name: 'Hobbytech GFX watch',
            keywords: ['GFX Camaro'],
            provider_scope: ['hobbytechtoys'],
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
      body: { error: 'unexpected_generic_scanner_run_for_hobbytech' },
    }).as('genericScannerRun')
    cy.intercept('POST', '/api/providers/hobbytech/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-hobbytech-live-1')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'hobbytechtoys',
          query_set_id: 'qs-hobbytech-live-1',
          candidates: [
            {
              id: 'cand-hobbytech-live-1',
              query_set_id: 'qs-hobbytech-live-1',
              listing_id: 'hobbytech-gfx-camaro',
              title: 'GFX Camaro Live Candidate',
              source: 'hobbytechtoys',
              price: 24.95,
              currency: 'AUD',
              url: 'https://hobbytechtoys.com.au/products/gfx-camaro',
              seller: 'hobbytechtoys.com.au',
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
    }).as('hobbytechProviderRun')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-hobbytech-live-1"]').click()
    cy.wait('@hobbytechProviderRun')
    cy.get('@genericScannerRun.all').should('have.length', 0)
    cy.get('[data-testid="scanner-action-status"]').should(
      'contain',
      'hobbytech_run_started_qs-hobbytech-live-1'
    )
    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-hobbytech-live-1"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('td', 'hobbytechtoys').should('be.visible')
      cy.contains('td', 'GFX Camaro Live Candidate').should('be.visible')
      cy.contains('td', '24.95 AUD').should('be.visible')
      cy.contains('td', 'https://hobbytechtoys.com.au/products/gfx-camaro').should(
        'be.visible'
      )
      cy.contains('td', 'in_stock').should('be.visible')
      cy.contains('td', 'wishlist_inventory_ready').should('be.visible')
    })
  })

  it('UC-AU-07 surfaces Hobbytech session-drift recovery after a bounded fallback retry', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-hobbytech-drift-1',
            name: 'Hobbytech drift watch',
            keywords: ['GFX Corvette'],
            provider_scope: ['hobbytechtoys'],
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
    cy.intercept('POST', '/api/providers/hobbytech/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-hobbytech-drift-1')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'hobbytechtoys',
          query_set_id: 'qs-hobbytech-drift-1',
          drift_recovered: true,
          warning: 'Recovered Hobbytech session drift with fallback discovery.',
          candidates: [
            {
              id: 'cand-hobbytech-drift-1',
              query_set_id: 'qs-hobbytech-drift-1',
              listing_id: 'hobbytech-gfx-corvette',
              title: 'GFX Corvette Recovered Candidate',
              source: 'hobbytechtoys',
              price: 34.95,
              currency: 'AUD',
              url: 'https://hobbytechtoys.com.au/products/gfx-corvette',
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
    }).as('hobbytechProviderRun')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-hobbytech-drift-1"]').click()
    cy.wait('@hobbytechProviderRun')
    cy.get('[data-testid="scanner-action-status"]').should(
      'contain',
      'hobbytech_run_started_qs-hobbytech-drift-1'
    )
    cy.get('[data-testid="scanner-action-feedback"]').should(
      'contain',
      'Recovered Hobbytech session drift with fallback discovery.'
    )
  })
})
