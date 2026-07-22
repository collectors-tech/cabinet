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

  it('UC-AU-04 preserves Frontline candidate provenance through Wishlist and Inventory handoff', () => {
    let wishlistEntries: Array<Record<string, unknown>> = []
    let wishlistItems: Array<Record<string, unknown>> = []
    let inventoryItems: Array<Record<string, unknown>> = []
    const frontlineHandoffAudit = {
      source: 'market_watch',
      source_provider: 'frontlinehobbies',
      query_set_id: 'qs-frontline-handoff-1',
      query_name: 'Frontline Handoff Watch',
      provider_scope: ['frontlinehobbies'],
      listing_id: 'frontline-afx-gt40',
      source_result_url: 'https://frontlinehobbies.com.au/products/afx-gt40',
      observed_price: 59.95,
      observed_currency: 'AUD',
      seller: 'frontlinehobbies.com.au',
    }

    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-frontline-handoff-1',
            name: 'Frontline Handoff Watch',
            keywords: ['AFX GT40'],
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
      expect(req.body.query_set_id).to.equal('qs-frontline-handoff-1')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'frontlinehobbies',
          query_set_id: 'qs-frontline-handoff-1',
          candidates: [
            {
              id: 'cand-frontline-handoff-1',
              query_set_id: 'qs-frontline-handoff-1',
              listing_id: 'frontline-afx-gt40',
              title: 'AFX GT40 Frontline Handoff Candidate',
              source: 'frontlinehobbies',
              price: 59.95,
              currency: 'AUD',
              url: 'https://frontlinehobbies.com.au/products/afx-gt40',
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
    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body.candidate_id).to.equal('cand-frontline-handoff-1')
      expect(req.body.payload).to.deep.equal({
        source: 'market_watch',
        query_set_id: 'qs-frontline-handoff-1',
      })
      if (req.body.type === 'add_to_wishlist') {
        wishlistEntries = [
          {
            id: 'wish-frontline-handoff-1',
            item_id: 'item-frontline-handoff-wishlist-1',
            priority: 'medium',
            target_price: 59.95,
            notes:
              'source_provider=frontlinehobbies; query_set_id=qs-frontline-handoff-1; query_name=Frontline Handoff Watch; provider_scope=frontlinehobbies',
            created_at: '2026-07-22T01:55:00Z',
            updated_at: '2026-07-22T01:55:00Z',
          },
        ]
        wishlistItems = [
          {
            id: 'item-frontline-handoff-wishlist-1',
            title: 'AFX GT40 Frontline Handoff Candidate',
            part_number: 'frontline-afx-gt40',
            status: 'wishlist',
            category: 'Slot Cars',
            priority: 'medium',
          },
        ]
        req.reply({
          statusCode: 200,
          body: {
            ok: true,
            action: 'add_to_wishlist',
            candidate_id: 'cand-frontline-handoff-1',
            audit: frontlineHandoffAudit,
          },
        })
        return
      }
      expect(req.body.type).to.equal('create_item')
      inventoryItems = [
        {
          id: 'item-frontline-handoff-inventory-1',
          title: 'AFX GT40 Frontline Handoff Candidate',
          part_number: 'frontline-afx-gt40',
          status: 'owned',
          category: 'Slot Cars',
          priority: 'medium',
          notes:
            '{"source_provider":"frontlinehobbies","query_set_id":"qs-frontline-handoff-1","query_name":"Frontline Handoff Watch","provider_scope":"frontlinehobbies","source_result_url":"https://frontlinehobbies.com.au/products/afx-gt40"}',
          description:
            '{"source_provider":"frontlinehobbies","query_set_id":"qs-frontline-handoff-1","query_name":"Frontline Handoff Watch","provider_scope":"frontlinehobbies","source_result_url":"https://frontlinehobbies.com.au/products/afx-gt40"}',
        },
      ]
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
          action: 'create_item',
          candidate_id: 'cand-frontline-handoff-1',
          audit: frontlineHandoffAudit,
        },
      })
    }).as('discoveryHandoff')
    cy.intercept('GET', '/api/wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } })
    }).as('wishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } })
    }).as('wishlistItems')
    cy.intercept('GET', '/api/items', (req) => {
      req.reply({ statusCode: 200, body: { items: inventoryItems } })
    }).as('inventoryItems')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })
    cy.intercept('GET', '/api/pricing/stats?item_id=item-frontline-handoff-wishlist-1', {
      statusCode: 200,
      body: { min: 59.95, median: 59.95, latest: 59.95 },
    }).as('wishlistPriceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=item-frontline-handoff-wishlist-1', {
      statusCode: 200,
      body: { points: [] },
    }).as('wishlistPriceTrend')
    cy.intercept('GET', '/api/pricing/history?item_id=item-frontline-handoff-wishlist-1', {
      statusCode: 200,
      body: { history: [] },
    }).as('wishlistPriceHistory')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-frontline-handoff-1"]').click()
    cy.wait('@frontlineProviderRun')
    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-frontline-handoff-1"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('td', 'frontlinehobbies').should('be.visible')
      cy.contains('td', 'AFX GT40 Frontline Handoff Candidate').should('be.visible')
      cy.contains('td', '59.95 AUD').should('be.visible')
      cy.contains('td', 'https://frontlinehobbies.com.au/products/afx-gt40').should(
        'be.visible'
      )
      cy.contains('td', 'wishlist_inventory_ready').should('be.visible')
    })

    cy.get('[data-testid="scanner-handoff-wishlist-qs-frontline-handoff-1"]').click()
    cy.wait('@discoveryHandoff').then((interception) => {
      expect(interception.response?.body).to.deep.include({
        ok: true,
        action: 'add_to_wishlist',
        candidate_id: 'cand-frontline-handoff-1',
      })
      expect(interception.response?.body.audit).to.deep.equal(frontlineHandoffAudit)
    })
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'wishlist_handoff_ok_cand-frontline-handoff-1'
    )

    cy.get('[data-testid="scanner-handoff-inventory-qs-frontline-handoff-1"]').click()
    cy.wait('@discoveryHandoff').then((interception) => {
      expect(interception.response?.body).to.deep.include({
        ok: true,
        action: 'create_item',
        candidate_id: 'cand-frontline-handoff-1',
      })
      expect(interception.response?.body.audit).to.deep.equal(frontlineHandoffAudit)
    })
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'inventory_handoff_ok_cand-frontline-handoff-1'
    )

    cy.visit('/wishlist/')
    cy.wait(['@wishlistEntries', '@wishlistItems'])
    cy.contains('AFX GT40 Frontline Handoff Candidate').should('be.visible')
    cy.contains('source_provider=frontlinehobbies').should('be.visible')
    cy.contains('query_set_id=qs-frontline-handoff-1').should('be.visible')
    cy.contains('provider_scope=frontlinehobbies').should('be.visible')

    cy.visit('/inventory/')
    cy.wait('@inventoryItems')
    cy.contains('button', 'Cards').click()
    cy.get('[data-testid="inventory-item-row-item-frontline-handoff-inventory-1"]')
      .should('contain', 'AFX GT40 Frontline Handoff Candidate')
    cy.get('[data-testid="inventory-card-notes-item-frontline-handoff-inventory-1"]')
      .should('be.visible')
      .and('contain', 'source_provider')
      .and('contain', 'frontlinehobbies')
      .and('contain', 'query_set_id')
      .and('contain', 'qs-frontline-handoff-1')
      .and('contain', 'source_result_url')
      .and('contain', 'https://frontlinehobbies.com.au/products/afx-gt40')
  })
})
