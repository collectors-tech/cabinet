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
})
