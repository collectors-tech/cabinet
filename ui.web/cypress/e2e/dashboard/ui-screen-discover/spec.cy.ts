describe('dashboard/ui-screen-discover', () => {
  function signInToDiscoveries() {
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true }).then((bootstrap) => {
      cy.request('PUT', '/api/profiles/active', { profile_id: bootstrap.profile_id })
        .its('status')
        .should('eq', 200)
      cy.visit('/sign-in?redirect=%2Fdiscoveries%2F', {
        onBeforeLoad(win) {
          win.localStorage.setItem(`cabinet.workspace.${bootstrap.profile_id}`, '1')
        },
      })
      cy.contains('button', 'Open local workspace').click()
      cy.get('body').then(($body) => {
        const profileButton = `Use ${bootstrap.profile_name}`
        if ($body.text().includes(profileButton)) {
          cy.contains('button', profileButton).click()
        }
      })
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/discoveries\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-DISCOVER-001 renders filterable candidate triage list', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', (req) => {
      if (req.query.q === 'camaro') {
        req.reply({
          statusCode: 200,
          body: {
            items: [
              {
                candidate_id: 'cand-001',
                title: 'AFX Camaro Wildfire',
                price: 42.5,
                url: 'https://example.test/camaro',
                last_seen: '2026-03-01T00:00:00Z',
                stock_state: 'in_stock',
                stock_count: 3,
              },
            ],
          },
        })
        return
      }
      req.reply({ statusCode: 200, body: { items: [] } })
    }).as('discoverList')

    signInToDiscoveries()
    cy.wait('@discoverList')

    cy.get('[data-testid="discover-filter-query"]').type('camaro')
    cy.get('[data-testid="discover-apply-filters"]').click()
    cy.wait('@discoverList')

    cy.get('[data-testid="discover-list"]').contains('AFX Camaro Wildfire').should('be.visible')
  })

  it('UI-SCREEN-DISCOVER-001 + UC-DIS-06 applies query price and date filters without route transition', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', (req) => {
      const hasFullFilterSet =
        req.query.q === 'mini-z' &&
        req.query.price_max === '90' &&
        req.query.date_from === '2026-06-01'
      if (hasFullFilterSet) {
        req.reply({
          statusCode: 200,
          body: {
            items: [
              {
                candidate_id: 'cand-filter-full',
                title: 'Kyosho Mini-Z Filter Candidate',
                price: 88.5,
                currency: 'AUD',
                url: 'https://example.test/mini-z-filter',
                last_seen: '2026-06-12T00:00:00Z',
                stock_state: 'in_stock',
                stock_count: 1,
              },
            ],
          },
        })
        return
      }
      req.reply({ statusCode: 200, body: { items: [] } })
    }).as('discoverFilteredList')

    signInToDiscoveries()
    cy.wait('@discoverFilteredList')

    cy.get('[data-testid="discover-filter-query"]').type('mini-z')
    cy.get('[data-testid="discover-filter-price"]').type('90')
    cy.get('[data-testid="discover-filter-date"]').type('2026-06-01')
    cy.get('[data-testid="discover-apply-filters"]').click()
    cy.wait('@discoverFilteredList')

    cy.location('pathname').should('match', /^\/discoveries\/?$/)
    cy.get('[data-testid="discover-list"]')
      .should('contain', 'Kyosho Mini-Z Filter Candidate')
      .and('contain', 'A$88.50')
  })

  it('UI-SCREEN-DISCOVER-003 shows loading state before candidate list resolves', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      delay: 750,
      body: {
        items: [
          {
            candidate_id: 'cand-loading-state',
            title: 'Loading State Candidate',
            price: 24.5,
            currency: 'AUD',
            url: 'https://example.test/loading-state',
            last_seen: '2026-06-12T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
          },
        ],
      },
    }).as('discoverListLoading')

    signInToDiscoveries()
    cy.get('[data-testid="discover-list"]').should(
      'contain',
      'Loading discoveries...'
    )
    cy.wait('@discoverListLoading')

    cy.location('pathname').should('match', /^\/discoveries\/?$/)
    cy.get('[data-testid="discover-list"]')
      .should('not.contain', 'Loading discoveries...')
      .and('contain', 'Loading State Candidate')
  })

  it('UI-SCREEN-DISCOVER-002 + UC-DIS-02..04 submits ignore wishlist track and create action payloads', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-002',
            title: 'Porsche GT3 Candidate',
            price: 31.0,
            url: 'https://example.test/porsche',
            last_seen: '2026-03-01T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 2,
          },
        ],
      },
    }).as('discoverList')

    const expectedActions = ['ignore', 'add_to_wishlist', 'track_price', 'create_item']
    cy.intercept('POST', '/api/discovery/action', (req) => {
      const expectedAction = expectedActions.shift()
      expect(req.body).to.deep.equal({
        candidate_id: 'cand-002',
        type: expectedAction,
        payload: {},
      })
      req.reply({ statusCode: 200, body: { ok: true } })
    }).as('discoverAction')

    signInToDiscoveries()
    cy.wait('@discoverList')

    cy.get('[data-testid="discover-action-ignore-cand-002"]').click()
    cy.wait('@discoverAction')
    cy.get('[data-testid="discover-action-status"]').should('contain', 'ignore:cand-002')

    cy.get('[data-testid="discover-action-wishlist-cand-002"]').click()
    cy.wait('@discoverAction')
    cy.get('[data-testid="discover-action-status"]').should('contain', 'add_to_wishlist:cand-002')

    cy.get('[data-testid="discover-action-track-cand-002"]').click()
    cy.wait('@discoverAction')
    cy.get('[data-testid="discover-action-status"]').should('contain', 'track_price:cand-002')

    cy.get('[data-testid="discover-action-create-cand-002"]').click()
    cy.wait('@discoverAction')
    cy.get('[data-testid="discover-action-status"]').should('contain', 'create_item:cand-002')
    cy.wrap(expectedActions).should('be.empty')
  })

  it('UI-SCREEN-DISCOVER-002 + UC-DIS-13 keeps candidate list stable when an action fails', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-action-failure',
            title: 'Action Failure Candidate',
            price: 33.75,
            currency: 'AUD',
            url: 'https://example.test/action-failure',
            last_seen: '2026-06-12T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
          },
        ],
      },
    }).as('discoverListFailureAction')

    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body).to.deep.equal({
        candidate_id: 'cand-action-failure',
        type: 'ignore',
        payload: {},
      })
      req.reply({ statusCode: 500, body: { error: 'action_failed' } })
    }).as('discoverActionFailure')

    signInToDiscoveries()
    cy.wait('@discoverListFailureAction')

    cy.get('[data-testid="discover-action-ignore-cand-action-failure"]').click()
    cy.wait('@discoverActionFailure')

    cy.location('pathname').should('match', /^\/discoveries\/?$/)
    cy.get('[data-testid="discover-action-status"]').should(
      'contain',
      'discover_action_500'
    )
    cy.get('[data-testid="discover-list"]').should(
      'contain',
      'Action Failure Candidate'
    )
    cy.get('@discoverListFailureAction.all').should('have.length', 1)
  })

  it('UI-SCREEN-DISCOVER-003 shows retryable error state when discover API fails', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 500,
      body: { error: 'failed_to_list_not_in_collection' },
    }).as('discoverListFailure')

    signInToDiscoveries()
    cy.wait('@discoverListFailure')
    cy.get('[data-testid="discover-error-state"]').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
  })

  it('UI-SCREEN-DISCOVER-004 keeps Discoveries as triage-only and excludes Market Watch query/run controls', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-003',
            title: 'AFX Boundary Candidate',
            price: 27.0,
            url: 'https://example.test/boundary',
            last_seen: '2026-03-01T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 4,
          },
        ],
      },
    }).as('discoverListBoundary')

    signInToDiscoveries()
    cy.wait('@discoverListBoundary')

    cy.get('[data-testid="discover-action-ignore-cand-003"]').should('be.visible')
    cy.get('[data-testid="discover-action-wishlist-cand-003"]').should('be.visible')
    cy.get('[data-testid="discover-action-track-cand-003"]').should('be.visible')
    cy.get('[data-testid="discover-action-create-cand-003"]').should('be.visible')

    cy.get('[data-testid="scanner-create-query"]').should('not.exist')
    cy.get('[data-testid^="scanner-run-"]').should('not.exist')
    cy.contains('Manage provider query sets').should('not.exist')
  })

  it('UI-SCREEN-DISCOVER-004 provides explicit handoff action to Market Watch with preserved context', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: { items: [] },
    }).as('discoverListHandoff')

    signInToDiscoveries()
    cy.wait('@discoverListHandoff')

    cy.get('[data-testid="discover-filter-query"]').type('afx')
    cy.get('[data-testid="discover-open-market-watch"]').click()

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
    cy.location('search').should('include', 'from=discoveries')
    cy.location('search').should('include', 'q=afx')
  })

  it('UI-SCREEN-DISCOVER-005 explains candidate inbox purpose', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: { items: [] },
    }).as('discoverListPurpose')

    signInToDiscoveries()
    cy.wait('@discoverListPurpose')

    cy.get('[data-testid="discover-candidate-inbox-purpose"]')
      .should('contain', 'Candidate inbox')
      .and('contain', 'Review found items')
      .and('contain', 'Wishlist')
      .and('contain', 'Inventory')
      .and('contain', 'Purchase')
      .and('contain', 'ignored or archived')

    cy.get('[data-testid="discover-list"]')
      .should('contain', 'No pending found-item candidates')
      .and('contain', 'separate from Inventory, Wishlist, and Market Watch query history')
  })

  it('UI-SCREEN-DISCOVER-005 + UC-DIS-14 renders empty candidate inbox without mutation controls', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: { items: [] },
    }).as('discoverListEmpty')

    signInToDiscoveries()
    cy.wait('@discoverListEmpty')

    cy.location('pathname').should('match', /^\/discoveries\/?$/)
    cy.get('[data-testid="discover-list"]')
      .should('be.visible')
      .and('contain', 'No pending found-item candidates')
      .and('contain', 'provider runs or imports')
    cy.get('[data-testid^="discover-candidate-row-"]').should('not.exist')
    cy.get('[data-testid^="discover-action-ignore-"]').should('not.exist')
    cy.get('[data-testid^="discover-action-wishlist-"]').should('not.exist')
    cy.get('[data-testid^="discover-action-track-"]').should('not.exist')
    cy.get('[data-testid^="discover-action-create-"]').should('not.exist')
    cy.get('[data-testid="discover-open-market-watch"]').should('be.visible')
    cy.location('pathname').should('match', /^\/discoveries\/?$/)
  })

  it('UI-SCREEN-DISCOVER-005 renders candidate provenance and destination actions', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-005',
            title: 'Kyosho Mini-Z Candidate',
            price: 88.5,
            currency: 'AUD',
            url: 'https://example.test/mini-z',
            source_result_url: 'https://provider.test/result/mini-z',
            source_result_id: 'result-555',
            provider: 'Hobbytech Toys',
            source: 'hobbytechtoys',
            first_seen: '2026-03-01T00:00:00Z',
            last_seen: '2026-03-08T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 2,
            triage_status: 'reviewing',
            confidence: 0.92,
            seller_label: 'Provider listing',
          },
        ],
      },
    }).as('discoverListProvenance')

    signInToDiscoveries()
    cy.wait('@discoverListProvenance')

    cy.get('[data-testid="discover-candidate-row-cand-005"]')
      .should('contain', 'Kyosho Mini-Z Candidate')
      .and('contain', 'A$88.50')
      .and('contain', 'Stock 2')

    cy.get('[data-testid="discover-provenance-cand-005"]')
      .should('contain', 'Hobbytech Toys')
      .and('contain', 'result-555')
      .and('contain', 'Provider listing')
    cy.get('[data-testid="discover-candidate-row-cand-005"]')
      .should('contain', 'First seen Mar 1, 2026')
      .and('contain', 'Last seen Mar 8, 2026')
      .and('contain', 'reviewing')
      .and('contain', 'Confidence 92%')

    cy.get('[data-testid="discover-action-review-source-cand-005"]')
      .should('have.attr', 'href', 'https://provider.test/result/mini-z')
      .and('have.attr', 'aria-label', 'Review source result')
    cy.get('[data-testid="discover-action-wishlist-cand-005"]').should(
      'have.attr',
      'aria-label',
      'Promote to Wishlist'
    )
    cy.get('[data-testid="discover-action-create-cand-005"]').should(
      'have.attr',
      'aria-label',
      'Inventory handoff'
    )
    cy.get('[data-testid="discover-action-track-cand-005"]').should(
      'have.attr',
      'aria-label',
      'Purchase follow-up'
    )
    cy.get('[data-testid="discover-action-ignore-cand-005"]').should(
      'have.attr',
      'aria-label',
      'Ignore or archive'
    )
  })

  it('UI-SCREEN-DISCOVER-007 + #1533 renders dashboard summary source filters and ranked deal table', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-deal-wishlist',
            title: 'AFX Wishlist Deal',
            price: 42,
            currency: 'AUD',
            target_price: 55,
            market_price_baseline: 70,
            price_delta_percent: 24,
            deal_score: 92,
            match_type: 'wishlist_match',
            match_reason: 'Wishlist match below target',
            wishlist_id: 'wish-001',
            source_provider: 'Market Watch',
            query_name: 'AFX saved search',
            source_result_url: 'https://provider.test/deal',
            listing_id: 'deal-001',
            seller_label: 'Hobby store',
            first_seen: '2026-06-20T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 2,
            triage_status: 'new',
            confidence: 0.96,
          },
          {
            candidate_id: 'cand-provider-attention',
            title: 'Provider Auth Candidate',
            price: 88,
            currency: 'AUD',
            match_type: 'provider_search',
            source_provider: 'Provider Store',
            source_trust_status: 'auth attention',
            first_seen: '2026-06-21T00:00:00Z',
            last_seen: '2026-06-24T00:00:00Z',
            stock_state: 'unknown',
            stock_count: 0,
            needs_review: true,
          },
          {
            candidate_id: 'cand-archived',
            title: 'Archived Candidate',
            price: 12,
            currency: 'AUD',
            match_type: 'market_watch_result',
            source_provider: 'Market Watch',
            first_seen: '2026-06-18T00:00:00Z',
            last_seen: '2026-06-19T00:00:00Z',
            stock_state: 'out_of_stock',
            stock_count: 0,
            triage_status: 'archived',
          },
        ],
      },
    }).as('discoverDashboardList')

    signInToDiscoveries()
    cy.wait('@discoverDashboardList')
      .its('request.query.include_archived')
      .should('eq', 'true')

    cy.get('[data-testid="discover-dashboard-summary"]')
      .should('contain', 'Best deals found')
      .and('contain', 'Wishlist matches')
      .and('contain', 'New since last visit')
      .and('contain', 'Market Watch review')
      .and('contain', 'Provider attention')

    cy.get('[data-testid="discover-list"] table').should('be.visible')
    cy.get('[data-testid="discover-candidate-row-cand-deal-wishlist"]')
      .should('contain', 'Wishlist match below target')
      .and('contain', 'A$42.00')
      .and('contain', 'Target A$55.00')
      .and('contain', 'Baseline A$70.00')
      .and('contain', '24% saving')
    cy.get('[data-testid="discover-candidate-row-cand-archived"]').should(
      'not.exist'
    )

    cy.get('[data-testid="discover-filter-tab-archived"]').click()
    cy.get('[data-testid="discover-candidate-row-cand-archived"]').should(
      'be.visible'
    )

    cy.get('[data-testid="discover-filter-tab-wishlist"]').click()
    cy.get('[data-testid="discover-candidate-row-cand-deal-wishlist"]').should(
      'be.visible'
    )
    cy.get('[data-testid="discover-candidate-row-cand-provider-attention"]').should(
      'not.exist'
    )
  })

  it('UI-SCREEN-DISCOVER-007 keeps ranking and source filters deterministic without mutating candidates', () => {
    cy.intercept('POST', '/api/discovery/action').as('unexpectedDiscoverAction')
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-provider-row',
            title: 'Provider Attention Row',
            price: 64,
            currency: 'AUD',
            match_type: 'provider_search',
            source_provider: 'Provider Store',
            source_trust_status: 'auth attention',
            first_seen: '2026-06-23T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'unknown',
            stock_count: 0,
            needs_review: true,
          },
          {
            candidate_id: 'cand-great-price',
            title: 'Great Price Candidate',
            price: 30,
            currency: 'AUD',
            target_price: 40,
            market_price_baseline: 60,
            deal_score: 95,
            match_type: 'store_stock',
            source_provider: 'Store Stock',
            first_seen: '2026-06-22T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 5,
            triage_status: 'new',
          },
          {
            candidate_id: 'cand-market-watch',
            title: 'Market Watch Candidate',
            price: 77,
            currency: 'AUD',
            match_type: 'market_watch_result',
            source_provider: 'Market Watch',
            query_name: 'Saved AFX search',
            first_seen: '2026-06-21T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
            triage_status: 'reviewing',
          },
          {
            candidate_id: 'cand-wishlist-first',
            title: 'Wishlist Ranked First',
            price: 42,
            currency: 'AUD',
            target_price: 55,
            market_price_baseline: 70,
            deal_score: 75,
            match_type: 'wishlist_match',
            match_reason: 'Wishlist match below target',
            wishlist_id: 'wish-ranked',
            source_provider: 'Market Watch',
            first_seen: '2026-06-20T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 2,
            triage_status: 'new',
            confidence: 0.98,
          },
          {
            candidate_id: 'cand-shared-inventory',
            title: 'Shared Inventory Candidate',
            price: 22,
            currency: 'AUD',
            match_type: 'public_binder_match',
            source_provider: 'Shared Binder',
            first_seen: '2026-06-19T00:00:00Z',
            last_seen: '2026-06-25T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
            triage_status: 'reviewing',
          },
          {
            candidate_id: 'cand-hidden-archived',
            title: 'Hidden Archived Candidate',
            price: 12,
            currency: 'AUD',
            match_type: 'market_watch_result',
            source_provider: 'Market Watch',
            first_seen: '2026-06-18T00:00:00Z',
            last_seen: '2026-06-24T00:00:00Z',
            stock_state: 'out_of_stock',
            stock_count: 0,
            triage_status: 'archived',
          },
        ],
      },
    }).as('discoverDashboardFilterList')

    signInToDiscoveries()
    cy.wait('@discoverDashboardFilterList')
      .its('request.query.include_archived')
      .should('eq', 'true')

    cy.get('[data-testid="discover-summary-best-deals-found"]').should(
      'contain',
      '2'
    )
    cy.get('[data-testid="discover-summary-wishlist-matches"]').should(
      'contain',
      '1'
    )
    cy.get('[data-testid="discover-summary-market-watch-review"]').should(
      'contain',
      '2'
    )
    cy.get('[data-testid="discover-summary-provider-attention"]').should(
      'contain',
      '1'
    )

    cy.get('[data-testid^="discover-candidate-row-"]').should('have.length', 5)
    cy.get('[data-testid^="discover-candidate-row-"]')
      .eq(0)
      .should('have.attr', 'data-testid', 'discover-candidate-row-cand-wishlist-first')
    cy.get('[data-testid^="discover-candidate-row-"]')
      .eq(1)
      .should('have.attr', 'data-testid', 'discover-candidate-row-cand-great-price')
    cy.get('[data-testid="discover-candidate-row-cand-hidden-archived"]').should(
      'not.exist'
    )

    cy.get('[data-testid="discover-filter-tab-market_watch"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid^="discover-candidate-row-"]').should('have.length', 2)
    cy.get('[data-testid="discover-candidate-row-cand-market-watch"]').should(
      'be.visible'
    )
    cy.get('[data-testid="discover-candidate-row-cand-hidden-archived"]').should(
      'not.exist'
    )

    cy.get('[data-testid="discover-filter-tab-stores"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid^="discover-candidate-row-"]').should('have.length', 2)
    cy.get('[data-testid="discover-candidate-row-cand-provider-row"]').should(
      'contain',
      'Provider Attention Row'
    )

    cy.get('[data-testid="discover-filter-tab-other"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid^="discover-candidate-row-"]').should('have.length', 1)
    cy.get('[data-testid="discover-candidate-row-cand-shared-inventory"]').should(
      'contain',
      'Shared Inventory Candidate'
    )

    cy.get('[data-testid="discover-filter-tab-archived"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid^="discover-candidate-row-"]').should('have.length', 1)
    cy.get('[data-testid="discover-candidate-row-cand-hidden-archived"]').should(
      'contain',
      'Hidden Archived Candidate'
    )
    cy.get('[data-testid="discover-action-status"]').should('not.exist')
    cy.get('@unexpectedDiscoverAction.all').should('have.length', 0)
  })

  it('UI-SCREEN-DISCOVER-007 sorts the dashboard table by deal and recency without mutating candidates', () => {
    cy.intercept('POST', '/api/discovery/action').as('unexpectedDiscoverAction')
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-middle-deal',
            title: 'Middle Deal Candidate',
            price: 40,
            currency: 'AUD',
            target_price: 45,
            market_price_baseline: 80,
            deal_score: 60,
            match_type: 'store_stock',
            source_provider: 'Store Stock',
            first_seen: '2026-06-19T00:00:00Z',
            last_seen: '2026-06-23T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
            triage_status: 'new',
          },
          {
            candidate_id: 'cand-best-deal',
            title: 'Best Deal Candidate',
            price: 25,
            currency: 'AUD',
            target_price: 35,
            market_price_baseline: 90,
            deal_score: 96,
            match_type: 'wishlist_match',
            source_provider: 'Market Watch',
            first_seen: '2026-06-18T00:00:00Z',
            last_seen: '2026-06-21T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 2,
            triage_status: 'new',
          },
          {
            candidate_id: 'cand-latest-seen',
            title: 'Latest Seen Candidate',
            price: 55,
            currency: 'AUD',
            target_price: 60,
            market_price_baseline: 70,
            deal_score: 35,
            match_type: 'market_watch_result',
            source_provider: 'Market Watch',
            first_seen: '2026-06-20T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 4,
            triage_status: 'reviewing',
          },
        ],
      },
    }).as('discoverSortableList')

    signInToDiscoveries()
    cy.wait('@discoverSortableList')

    cy.get('[data-testid^="discover-candidate-row-"]')
      .eq(0)
      .should('have.attr', 'data-testid', 'discover-candidate-row-cand-best-deal')

    cy.get('[data-testid="discover-sort-last-seen"]').click()
    cy.get('[data-testid^="discover-candidate-row-"]')
      .eq(0)
      .should(
        'have.attr',
        'data-testid',
        'discover-candidate-row-cand-latest-seen'
      )

    cy.get('[data-testid="discover-sort-deal"]').click()
    cy.get('[data-testid^="discover-candidate-row-"]')
      .eq(0)
      .should('have.attr', 'data-testid', 'discover-candidate-row-cand-best-deal')
    cy.get('[data-testid="discover-action-status"]').should('not.exist')
    cy.get('@unexpectedDiscoverAction.all').should('have.length', 0)
  })

  it('UI-SCREEN-DISCOVER-007 renders provider-attention no-match and promoted destination states', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-provider-needs-attention',
            title: 'Provider Attention Candidate',
            price: 88,
            currency: 'AUD',
            match_type: 'provider_search',
            source_provider: 'Provider Store',
            source_trust_status: 'auth attention',
            first_seen: '2026-06-21T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'unknown',
            stock_count: 0,
            needs_review: true,
          },
          {
            candidate_id: 'cand-no-match',
            title: 'Unmatched Source Candidate',
            price: 18,
            currency: 'AUD',
            match_type: 'manual_capture',
            source_label: 'Manual Capture',
            first_seen: '2026-06-20T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
            triage_status: 'new',
          },
          {
            candidate_id: 'cand-already-promoted',
            title: 'Already Promoted Candidate',
            price: 44,
            currency: 'AUD',
            match_type: 'wishlist_match',
            source_provider: 'Market Watch',
            first_seen: '2026-06-19T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
            triage_status: 'wishlisted',
            destination_status: 'wishlisted',
            destination_link: '/wishlist/?candidate=cand-already-promoted',
          },
        ],
      },
    }).as('discoverDashboardStateList')

    signInToDiscoveries()
    cy.wait('@discoverDashboardStateList')

    cy.get('[data-testid="discover-summary-provider-attention"]').should(
      'contain',
      '1'
    )
    cy.get('[data-testid="discover-candidate-row-cand-provider-needs-attention"]')
      .should('contain', 'Provider Store')
      .and('contain', 'auth attention')
      .and('contain', 'Needs review')

    cy.get('[data-testid="discover-candidate-row-cand-no-match"]')
      .should('contain', 'Found candidate')
      .and('contain', 'Review ready')
    cy.get('[data-testid="discover-filter-tab-wishlist"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid="discover-candidate-row-cand-no-match"]').should(
      'not.exist'
    )

    cy.get('[data-testid="discover-filter-tab-all"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid="discover-action-open-destination-cand-already-promoted"]')
      .should('have.attr', 'href', '/wishlist/?candidate=cand-already-promoted')
      .and('contain', 'Open destination')
    cy.get('[data-testid="discover-action-wishlist-cand-already-promoted"]').should(
      'not.exist'
    )
    cy.get('[data-testid="discover-action-create-cand-already-promoted"]').should(
      'not.exist'
    )
  })

  it('UI-SCREEN-DISCOVER-007 + #1556 shows contextual actions for wishlist, new, and archived candidates', () => {
    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-wishlist-match-actions',
            title: 'Wishlist Match Action Candidate',
            price: 35,
            currency: 'AUD',
            match_type: 'wishlist_match',
            wishlist_id: 'wish-action-1',
            wishlist_item_id: 'item-action-1',
            source_result_url: 'https://provider.test/wishlist-match',
            first_seen: '2026-06-21T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
            triage_status: 'new',
          },
          {
            candidate_id: 'cand-new-actions',
            title: 'New Discovery Action Candidate',
            price: 22,
            currency: 'AUD',
            match_type: 'provider_search',
            source_result_url: 'https://provider.test/new-result',
            first_seen: '2026-06-21T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 3,
            triage_status: 'new',
          },
          {
            candidate_id: 'cand-archived-actions',
            title: 'Archived Action Candidate',
            price: 18,
            currency: 'AUD',
            match_type: 'provider_search',
            source_result_url: 'https://provider.test/archived-result',
            first_seen: '2026-06-20T00:00:00Z',
            last_seen: '2026-06-26T00:00:00Z',
            stock_state: 'out_of_stock',
            stock_count: 0,
            triage_status: 'archived',
          },
        ],
      },
    }).as('discoverContextActionsList')

    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body).to.deep.equal({
        candidate_id: 'cand-archived-actions',
        type: 'review',
        payload: {},
      })
      req.reply({ statusCode: 200, body: { ok: true } })
    }).as('discoverRestoreAction')

    signInToDiscoveries()
    cy.wait('@discoverContextActionsList')

    cy.get('[data-testid="discover-candidate-row-cand-wishlist-match-actions"]')
      .should('contain', 'Wishlist match')
      .within(() => {
        cy.get('[data-testid="discover-action-review-source-cand-wishlist-match-actions"]').should('exist')
        cy.get('[data-testid="discover-action-track-cand-wishlist-match-actions"]').should('exist')
        cy.get('[data-testid="discover-action-ignore-cand-wishlist-match-actions"]').should('exist')
        cy.get('[data-testid="discover-action-wishlist-cand-wishlist-match-actions"]').should('not.exist')
        cy.get('[data-testid="discover-action-create-cand-wishlist-match-actions"]').should('not.exist')
      })

    cy.get('[data-testid="discover-candidate-row-cand-new-actions"]').within(() => {
      cy.get('[data-testid="discover-action-review-source-cand-new-actions"]').should('exist')
      cy.get('[data-testid="discover-action-wishlist-cand-new-actions"]').should('exist')
      cy.get('[data-testid="discover-action-track-cand-new-actions"]').should('exist')
      cy.get('[data-testid="discover-action-create-cand-new-actions"]').should('exist')
      cy.get('[data-testid="discover-action-ignore-cand-new-actions"]').should('exist')
    })

    cy.get('[data-testid="discover-filter-tab-archived"]').click()
    cy.get('[data-testid="discover-candidate-row-cand-archived-actions"]').within(() => {
      cy.get('[data-testid="discover-action-restore-cand-archived-actions"]').should('exist')
      cy.get('[data-testid="discover-action-wishlist-cand-archived-actions"]').should('not.exist')
      cy.get('[data-testid="discover-action-track-cand-archived-actions"]').should('not.exist')
      cy.get('[data-testid="discover-action-create-cand-archived-actions"]').should('not.exist')
    })

    cy.get('[data-testid="discover-action-restore-cand-archived-actions"]').click()
    cy.wait('@discoverRestoreAction')
    cy.get('[data-testid="discover-action-status"]').should(
      'contain',
      'review:cand-archived-actions'
    )
  })

  it('UI-SCREEN-DISCOVER-006 promotes a candidate to Wishlist without purchased state', () => {
    let wishlistEntries: Array<Record<string, unknown>> = []
    let wishlistItems: Array<Record<string, unknown>> = []

    cy.intercept('GET', '/api/discovery/not-in-collection*', {
      statusCode: 200,
      body: {
        items: [
          {
            candidate_id: 'cand-wishlist-promotion',
            title: 'Discovery Wishlist Promotion Candidate',
            price: 52.25,
            currency: 'AUD',
            url: 'https://example.test/promotion',
            source_result_url: 'https://provider.test/result/promotion',
            source_result_id: 'result-promotion-001',
            provider: 'Provider Search',
            source: 'provider-search',
            first_seen: '2026-06-01T00:00:00Z',
            last_seen: '2026-06-12T00:00:00Z',
            stock_state: 'in_stock',
            stock_count: 1,
            triage_status: 'new',
            confidence: 0.88,
            seller_label: 'Provider listing',
          },
        ],
      },
    }).as('discoverListPromotion')

    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body).to.deep.include({
        candidate_id: 'cand-wishlist-promotion',
        type: 'add_to_wishlist',
      })
      wishlistItems = [
        {
          id: 'item-promoted-1',
          title: 'Discovery Wishlist Promotion Candidate',
          part_number: 'PROMO-001',
          status: 'wishlist',
          category: 'Promotions',
          priority: 'medium',
        },
      ]
      wishlistEntries = [
        {
          id: 'wish-promoted-1',
          item_id: 'item-promoted-1',
          priority: 'medium',
          target_price: 52.25,
          notes: 'Promoted from Discoveries result-promotion-001',
          owned: false,
          delivered: false,
          quantity: 0,
          needed_quantity: 1,
        },
      ]
      req.reply({ statusCode: 200, body: { ok: true } })
    }).as('promoteToWishlist')
    cy.intercept('GET', '/api/wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } })
    }).as('wishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } })
    }).as('wishlistItems')
    cy.intercept('GET', '/api/pricing/stats?item_id=item-promoted-1', {
      statusCode: 200,
      body: { latest: 52.25 },
    }).as('promotedPriceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=item-promoted-1', {
      statusCode: 200,
      body: { points: [] },
    }).as('promotedPriceTrend')

    signInToDiscoveries()
    cy.wait('@discoverListPromotion')

    cy.get('[data-testid="discover-action-wishlist-cand-wishlist-promotion"]').click()
    cy.wait('@promoteToWishlist')
    cy.get('[data-testid="discover-action-status"]').should(
      'contain',
      'add_to_wishlist:cand-wishlist-promotion'
    )

    cy.visit('/wishlist/')
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')
    cy.wait('@promotedPriceStats')
    cy.wait('@promotedPriceTrend')

    cy.get('button[aria-label="Switch to rows view"]').click()
    cy.contains('tr', 'Discovery Wishlist Promotion Candidate').within(() => {
      cy.get('[data-testid="wishlist-category-item-promoted-1"]').should(
        'contain.text',
        'Promotions'
      )
      cy.get('[data-testid="wishlist-purchase-open-item-promoted-1"]').should(
        'be.visible'
      )
      cy.get('[data-testid="wishlist-delivered-checkbox-item-promoted-1"]').should(
        'not.exist'
      )
      cy.contains('Promoted from Discoveries result-promotion-001').should('be.visible')
    })

    cy.contains('button', 'Cards').click()
    cy.get('[data-testid="wishlist-card-purchased-item-promoted-1"]').should(
      'contain.text',
      'Purchased: No'
    )
    cy.get('[data-testid="wishlist-card-delivered-item-promoted-1"]').should(
      'contain.text',
      'Delivered: No'
    )
    cy.contains('Category: Promotions').should('be.visible')
  })
})
