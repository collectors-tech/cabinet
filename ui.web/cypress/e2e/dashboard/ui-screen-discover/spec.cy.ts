describe('dashboard/ui-screen-discover', () => {
  function signInToDiscoveries() {
    cy.visit('/sign-in?redirect=%2Fdiscoveries%2F')
    cy.get('input[name="email"]').clear().type('e2e-discover@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
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

  it('UI-SCREEN-DISCOVER-002 executes ignore/wishlist/track/create actions deterministically', () => {
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

    cy.intercept('POST', '/api/discovery/action', { statusCode: 200, body: { ok: true } }).as(
      'discoverAction'
    )

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
      .and('contain', 'First seen Mar 1, 2026; last seen Mar 8, 2026')
      .and('contain', 'reviewing - Confidence 92%')
      .and('contain', 'Provider listing')

    cy.get('[data-testid="discover-source-result-cand-005"]')
      .should('have.attr', 'href', 'https://provider.test/result/mini-z')
      .and('contain', 'Review source result')
    cy.get('[data-testid="discover-action-review-source-cand-005"]').should('be.visible')
    cy.get('[data-testid="discover-action-wishlist-cand-005"]').should(
      'contain',
      'Promote to Wishlist'
    )
    cy.get('[data-testid="discover-action-create-cand-005"]').should(
      'contain',
      'Inventory Handoff'
    )
    cy.get('[data-testid="discover-action-track-cand-005"]').should(
      'contain',
      'Purchase Follow-up'
    )
    cy.get('[data-testid="discover-action-ignore-cand-005"]').should(
      'contain',
      'Ignore / Archive'
    )
  })
})
