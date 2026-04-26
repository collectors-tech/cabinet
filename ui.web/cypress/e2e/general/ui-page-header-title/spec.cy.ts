describe('ui-page-header-title', () => {
  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-header-title@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', /^Sign in$/).click()
  }

  function assertCenteredHeader(testId: string, title: string) {
    cy.get(`[data-testid="${testId}-header-title"]`)
      .should('be.visible')
      .and('have.attr', 'data-centered', 'true')
      .and('contain', title)
    cy.get(`[data-testid="${testId}-page-icon"]`).should('be.visible')
    cy.get('header').should('not.contain', 'Active:')
    cy.get('header').should('not.contain', 'Collection:')
    cy.get('header').should('not.contain', 'Planning list')
  }

  function assertHeaderTitleDoesNotOverlapActions(testId: string) {
    cy.get(`[data-testid="${testId}-header-title"]`).then(($title) => {
      const titleRect = $title[0].getBoundingClientRect()

      cy.get(`[data-testid="${testId}-global-header-actions"]`).then(
        ($actions) => {
          const actionsRect = $actions[0].getBoundingClientRect()

          expect(
            titleRect.right,
            `${testId} title stays clear of header actions`
          ).to.be.lessThan(actionsRect.left - 8)
        }
      )
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-PAGE-HEADER-TITLE-001 hides crowded titles instead of overlapping header actions', () => {
    cy.viewport(1240, 720)

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get('[data-testid="inventory-header-title"]')
      .should('have.attr', 'data-crowded', 'true')
      .and('not.be.visible')
    cy.get('[data-testid="inventory-global-header-actions"]').should('be.visible')
    cy.get('header').should('not.contain', 'Active:')
    cy.get('header').should('not.contain', 'Collection:')
    cy.get('header').should('not.contain', 'Planning list')
  })

  it('UI-PAGE-HEADER-TITLE-002 centers primary page titles with icons and no inline context text', () => {
    cy.viewport(2048, 900)

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    assertCenteredHeader('inventory', 'Inventory')
    assertHeaderTitleDoesNotOverlapActions('inventory')

    cy.visit('/collections/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/collections\/?$/)
    assertCenteredHeader('collections', 'Collections')
    assertHeaderTitleDoesNotOverlapActions('collections')

    cy.visit('/wishlist/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    assertCenteredHeader('wishlist', 'Wishlist')
    assertHeaderTitleDoesNotOverlapActions('wishlist')
  })
})
