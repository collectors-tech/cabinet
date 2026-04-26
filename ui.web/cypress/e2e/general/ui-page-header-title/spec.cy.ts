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

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.viewport(1440, 900)
  })

  it('UI-PAGE-HEADER-TITLE-001 centers primary page titles with icons and no inline context text', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    assertCenteredHeader('inventory', 'Inventory')

    cy.visit('/collections/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/collections\/?$/)
    assertCenteredHeader('collections', 'Collections')

    cy.visit('/wishlist/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    assertCenteredHeader('wishlist', 'Wishlist')
  })
})
