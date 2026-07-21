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

  it('UC-AU-06 preserves Hobbytech candidate provenance through Wishlist and Inventory handoff', () => {
    let wishlistEntries: Array<Record<string, unknown>> = []
    let wishlistItems: Array<Record<string, unknown>> = []
    let inventoryItems: Array<Record<string, unknown>> = []
    const hobbytechHandoffAudit = {
      source: 'market_watch',
      source_provider: 'hobbytechtoys',
      query_set_id: 'qs-hobbytech-handoff-1',
      query_name: 'Hobbytech Handoff Watch',
      provider_scope: ['hobbytechtoys'],
      listing_id: 'hobbytech-gfx-mustang',
      source_result_url: 'https://hobbytechtoys.com.au/products/gfx-mustang',
      observed_price: 29.95,
      observed_currency: 'AUD',
      seller: 'hobbytechtoys.com.au',
    }

    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-hobbytech-handoff-1',
            name: 'Hobbytech Handoff Watch',
            keywords: ['GFX Mustang'],
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
      expect(req.body.query_set_id).to.equal('qs-hobbytech-handoff-1')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'hobbytechtoys',
          query_set_id: 'qs-hobbytech-handoff-1',
          candidates: [
            {
              id: 'cand-hobbytech-handoff-1',
              query_set_id: 'qs-hobbytech-handoff-1',
              listing_id: 'hobbytech-gfx-mustang',
              title: 'GFX Mustang Handoff Candidate',
              source: 'hobbytechtoys',
              price: 29.95,
              currency: 'AUD',
              url: 'https://hobbytechtoys.com.au/products/gfx-mustang',
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
    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body.candidate_id).to.equal('cand-hobbytech-handoff-1')
      expect(req.body.payload).to.deep.equal({
        source: 'market_watch',
        query_set_id: 'qs-hobbytech-handoff-1',
      })
      if (req.body.type === 'add_to_wishlist') {
        wishlistEntries = [
          {
            id: 'wish-hobbytech-handoff-1',
            item_id: 'item-hobbytech-handoff-wishlist-1',
            priority: 'medium',
            target_price: 29.95,
            notes:
              'source_provider=hobbytechtoys; query_set_id=qs-hobbytech-handoff-1; query_name=Hobbytech Handoff Watch; provider_scope=hobbytechtoys',
            created_at: '2026-07-18T17:51:00Z',
            updated_at: '2026-07-18T17:51:00Z',
          },
        ]
        wishlistItems = [
          {
            id: 'item-hobbytech-handoff-wishlist-1',
            title: 'GFX Mustang Handoff Candidate',
            part_number: 'hobbytech-gfx-mustang',
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
            candidate_id: 'cand-hobbytech-handoff-1',
            audit: hobbytechHandoffAudit,
          },
        })
        return
      }
      expect(req.body.type).to.equal('create_item')
      inventoryItems = [
        {
          id: 'item-hobbytech-handoff-inventory-1',
          title: 'GFX Mustang Handoff Candidate',
          part_number: 'hobbytech-gfx-mustang',
          status: 'owned',
          category: 'Slot Cars',
          priority: 'medium',
          notes:
            '{"source_provider":"hobbytechtoys","query_set_id":"qs-hobbytech-handoff-1","query_name":"Hobbytech Handoff Watch","provider_scope":"hobbytechtoys","source_result_url":"https://hobbytechtoys.com.au/products/gfx-mustang"}',
          description:
            '{"source_provider":"hobbytechtoys","query_set_id":"qs-hobbytech-handoff-1","query_name":"Hobbytech Handoff Watch","provider_scope":"hobbytechtoys","source_result_url":"https://hobbytechtoys.com.au/products/gfx-mustang"}',
        },
      ]
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
          action: 'create_item',
          candidate_id: 'cand-hobbytech-handoff-1',
          audit: hobbytechHandoffAudit,
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
    cy.intercept('GET', '/api/pricing/stats?item_id=item-hobbytech-handoff-wishlist-1', {
      statusCode: 200,
      body: { min: 29.95, median: 29.95, latest: 29.95 },
    }).as('wishlistPriceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=item-hobbytech-handoff-wishlist-1', {
      statusCode: 200,
      body: { points: [] },
    }).as('wishlistPriceTrend')
    cy.intercept('GET', '/api/pricing/history?item_id=item-hobbytech-handoff-wishlist-1', {
      statusCode: 200,
      body: { history: [] },
    }).as('wishlistPriceHistory')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-hobbytech-handoff-1"]').click()
    cy.wait('@hobbytechProviderRun')
    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-hobbytech-handoff-1"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('td', 'hobbytechtoys').should('be.visible')
      cy.contains('td', 'GFX Mustang Handoff Candidate').should('be.visible')
      cy.contains('td', '29.95 AUD').should('be.visible')
      cy.contains('td', 'https://hobbytechtoys.com.au/products/gfx-mustang').should(
        'be.visible'
      )
      cy.contains('td', 'wishlist_inventory_ready').should('be.visible')
    })

    cy.get('[data-testid="scanner-handoff-wishlist-qs-hobbytech-handoff-1"]').click()
    cy.wait('@discoveryHandoff').then((interception) => {
      expect(interception.response?.body).to.deep.include({
        ok: true,
        action: 'add_to_wishlist',
        candidate_id: 'cand-hobbytech-handoff-1',
      })
      expect(interception.response?.body.audit).to.deep.equal(hobbytechHandoffAudit)
    })
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'wishlist_handoff_ok_cand-hobbytech-handoff-1'
    )

    cy.get('[data-testid="scanner-handoff-inventory-qs-hobbytech-handoff-1"]').click()
    cy.wait('@discoveryHandoff').then((interception) => {
      expect(interception.response?.body).to.deep.include({
        ok: true,
        action: 'create_item',
        candidate_id: 'cand-hobbytech-handoff-1',
      })
      expect(interception.response?.body.audit).to.deep.equal(hobbytechHandoffAudit)
    })
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'inventory_handoff_ok_cand-hobbytech-handoff-1'
    )

    cy.visit('/wishlist/')
    cy.wait(['@wishlistEntries', '@wishlistItems'])
    cy.contains('GFX Mustang Handoff Candidate').should('be.visible')
    cy.contains('source_provider=hobbytechtoys').should('be.visible')
    cy.contains('query_set_id=qs-hobbytech-handoff-1').should('be.visible')
    cy.contains('provider_scope=hobbytechtoys').should('be.visible')

    cy.visit('/inventory/')
    cy.wait('@inventoryItems')
    cy.contains('button', 'Cards').click()
    cy.get('[data-testid="inventory-item-row-item-hobbytech-handoff-inventory-1"]')
      .should('contain', 'GFX Mustang Handoff Candidate')
    cy.get('[data-testid="inventory-card-notes-item-hobbytech-handoff-inventory-1"]')
      .should('be.visible')
      .and('contain', 'source_provider')
      .and('contain', 'hobbytechtoys')
      .and('contain', 'query_set_id')
      .and('contain', 'qs-hobbytech-handoff-1')
      .and('contain', 'source_result_url')
      .and('contain', 'https://hobbytechtoys.com.au/products/gfx-mustang')
  })
})
