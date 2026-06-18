describe('general/icon-only-navigation', () => {
  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-icon-nav@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-018 renders primary navigation as icon-only accessible controls', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    const iconOnlyLinks = [
      ['dashboard', 'Dashboard'],
      ['inventory', 'Inventory'],
      ['media', 'Media'],
      ['collections', 'Collections'],
      ['wishlist', 'Wishlist'],
      ['discoveries', 'Discoveries'],
      ['market-watch', 'Market Watch'],
      ['purchases', 'Purchases'],
      ['integrations', 'Integrations'],
      ['chats', 'Chats'],
      ['users', 'Users'],
      ['reports', 'Reports'],
    ] as const

    iconOnlyLinks.forEach(([key, label]) => {
      cy.get(`[data-testid="sidebar-nav-link-${key}"]`)
        .scrollIntoView()
        .should('be.visible')
        .and('have.attr', 'aria-label', label)
        .within(() => {
          cy.get('svg').should('be.visible')
          cy.get(`[data-testid="sidebar-nav-label-${key}"]`).should('not.exist')
        })
    })

    cy.get('[data-testid="sidebar-nav-link-inventory"]')
      .should('have.attr', 'data-active', 'true')
      .focus()
      .should('be.focused')
  })
})
