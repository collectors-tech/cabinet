describe('integrations/default-site-search', () => {
  function signInToMarketWatch() {
    cy.visit('/sign-in?redirect=%2Fscanner%2F')
    cy.get('input[name="email"]').clear().type('e2e-default-search@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('DEFAULT-SITE-SEARCH-004 manages provider-bound saved searches with persisted filters', () => {
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', { statusCode: 200, body: { status: 'ok' } }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/query-sets', (req) => {
      expect(req.body.provider_scope).to.deep.equal(['bonzaslotcars'])
      expect(req.body.keywords).to.deep.equal(['afx', 'mega g+'])
      expect(req.body.schedule_cron).to.equal('0 */4 * * *')
      req.reply({
        statusCode: 201,
        body: {
          id: 'qs-dss-1',
          name: 'Bonza AFX',
          keywords: ['afx', 'mega g+'],
          provider_scope: ['bonzaslotcars'],
          schedule_cron: '0 */4 * * *',
          enabled: true,
        },
      })
    }).as('createQuerySet')
    cy.intercept('PUT', '/api/scanner/query-sets/qs-dss-1', (req) => {
      expect(req.body.name).to.equal('Bonza AFX Updated')
      expect(req.body.keywords).to.deep.equal(['afx', 'mega g+', 'camaro'])
      expect(req.body.schedule_cron).to.equal('0 */8 * * *')
      req.reply({
        statusCode: 200,
        body: {
          ...req.body,
          id: 'qs-dss-1',
          provider_scope: ['bonzaslotcars'],
          enabled: true,
        },
      })
    }).as('updateQuerySet')
    cy.intercept('DELETE', '/api/scanner/query-sets/qs-dss-1', { statusCode: 204, body: '' }).as('deleteQuerySet')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-provider-single"]').select('bonzaslotcars')
    cy.get('[data-testid="scanner-new-query-name"]').type('Bonza AFX')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('afx, mega g+')
    cy.get('[data-testid="scanner-new-query-schedule"]').clear().type('0 */4 * * *')
    cy.get('[data-testid="scanner-create-query"]').click()
    cy.wait('@createQuerySet')

    cy.get('[data-testid="scanner-edit-qs-dss-1"]').click()
    cy.get('[data-testid="scanner-edit-name-qs-dss-1"]').clear().type('Bonza AFX Updated')
    cy.get('[data-testid="scanner-edit-keywords-qs-dss-1"]').clear().type('afx, mega g+, camaro')
    cy.get('[data-testid="scanner-edit-schedule-qs-dss-1"]').clear().type('0 */8 * * *')
    cy.get('[data-testid="scanner-save-qs-dss-1"]').click()
    cy.wait('@updateQuerySet')
    cy.contains('[data-testid="scanner-query-list"]', 'Bonza AFX Updated').should('be.visible')

    cy.get('[data-testid="scanner-delete-qs-dss-1"]').click()
    cy.wait('@deleteQuerySet')
    cy.get('[data-testid="scanner-action-status"]').should('contain', 'query_set_deleted_qs-dss-1')
  })

  it('DEFAULT-SITE-SEARCH-005 runs saved searches now and through scheduled refresh', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-dss-2',
            name: 'Amazon Watch',
            keywords: ['slot cars'],
            provider_scope: ['amazon'],
            schedule_cron: '0 */6 * * *',
            enabled: true,
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', { statusCode: 200, body: { status: 'ok' } }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', {
      statusCode: 200,
      body: { run_id: 'run-dss-2', status: 'ok' },
    }).as('runNow')
    cy.intercept('GET', '/api/scanner/candidates?query_set_id=qs-dss-2', {
      statusCode: 200,
      body: {
        candidates: [
          {
            id: 'cand-dss-2',
            query_set_id: 'qs-dss-2',
            listing_id: 'amazon-1',
            title: 'Amazon Slot Car',
            source: 'amazon',
          },
        ],
      },
    }).as('candidates')
    cy.intercept('POST', '/api/scanner/run/scheduled', {
      statusCode: 200,
      body: {
        run_id: 'scheduled-1',
        query_sets_executed: 1,
        candidates_collected: 1,
        failures: 0,
      },
    }).as('scheduledRefresh')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-dss-2"]').click()
    cy.wait('@runNow')
    cy.wait('@candidates')
    cy.get('[data-testid="scanner-action-status"]').should('contain', 'run_started_qs-dss-2')
    cy.get('[data-testid="scanner-candidates-qs-dss-2"]').should('contain', 'Amazon Slot Car')

    cy.get('[data-testid="scanner-run-scheduled-refresh"]').click()
    cy.wait('@scheduledRefresh')
    cy.get('[data-testid="scanner-action-status"]').should('contain', 'scheduled_run_completed_scheduled-1')
    cy.get('[data-testid="scanner-action-feedback"]').should('contain', 'Query sets executed: 1')
  })

  it('DEFAULT-SITE-SEARCH-006 hands off saved-search output to discoveries and persisted wishlist flows', () => {
    let wishlistEntries: Array<Record<string, unknown>> = []
    let wishlistItems: Array<Record<string, unknown>> = []

    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-dss-3',
            name: 'eBay Handoff',
            keywords: ['ho slot'],
            provider_scope: ['ebay'],
            enabled: true,
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', { statusCode: 200, body: { status: 'ok' } }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', {
      statusCode: 200,
      body: { run_id: 'run-dss-3', status: 'ok' },
    }).as('runNow')
    cy.intercept('GET', '/api/scanner/candidates?query_set_id=qs-dss-3', {
      statusCode: 200,
      body: {
        candidates: [
          {
            id: 'cand-dss-3',
            query_set_id: 'qs-dss-3',
            listing_id: 'ebay-1',
            title: 'eBay Handoff Car',
            source: 'ebay',
          },
        ],
      },
    }).as('candidates')
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [{ candidate_id: 'cand-dss-3', title: 'eBay Handoff Car' }],
      },
    }).as('discoveriesHandoff')
    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body.candidate_id).to.equal('cand-dss-3')
      expect(req.body.type).to.equal('add_to_wishlist')
      expect(req.body.payload).to.deep.equal({
        source: 'market_watch',
        query_set_id: 'qs-dss-3',
      })
      wishlistEntries = [
        {
          id: 'wish-dss-3',
          item_id: 'item-dss-3',
          priority: 'medium',
          target_price: 0,
          notes:
            'source_provider=ebay; query_set_id=qs-dss-3; query_name=eBay Handoff; provider_scope=ebay',
          created_at: '2026-05-27T10:44:00Z',
          updated_at: '2026-05-27T10:44:00Z',
        },
      ]
      wishlistItems = [
        {
          id: 'item-dss-3',
          title: 'eBay Handoff Car',
          part_number: 'ebay-1',
          status: 'wishlist',
          category: 'Slot Cars',
          priority: 'medium',
        },
      ]
      req.reply({ statusCode: 200, body: { ok: true } })
    }).as('wishlistHandoff')
    cy.intercept('GET', '/api/wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } })
    }).as('wishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } })
    }).as('wishlistItems')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })
    cy.intercept('GET', '/api/pricing/stats?item_id=item-dss-3', {
      statusCode: 200,
      body: { min: 0, median: 0, latest: 0 },
    }).as('wishlistPriceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=item-dss-3', {
      statusCode: 200,
      body: { points: [] },
    }).as('wishlistPriceTrend')
    cy.intercept('GET', '/api/pricing/history?item_id=item-dss-3', {
      statusCode: 200,
      body: { history: [] },
    }).as('wishlistPriceHistory')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-dss-3"]').click()
    cy.wait('@runNow')
    cy.wait('@candidates')

    cy.get('[data-testid="scanner-handoff-discoveries-qs-dss-3"]').click()
    cy.wait('@discoveriesHandoff')
    cy.get('[data-testid="scanner-handoff-status"]').should('contain', 'discoveries_handoff_ok_1')

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-dss-3"]').click()
    cy.get('[data-testid="scanner-handoff-wishlist-qs-dss-3"]').click()
    cy.wait('@wishlistHandoff')
    cy.get('[data-testid="scanner-handoff-status"]').should('contain', 'wishlist_handoff_ok')

    cy.visit('/wishlist/')
    cy.wait(['@wishlistEntries', '@wishlistItems'])
    cy.contains('eBay Handoff Car').should('be.visible')
    cy.contains('source_provider=ebay').should('be.visible')
    cy.contains('query_set_id=qs-dss-3').should('be.visible')
    cy.contains('query_name=eBay Handoff').should('be.visible')
    cy.contains('provider_scope=ebay').should('be.visible')
  })
})
