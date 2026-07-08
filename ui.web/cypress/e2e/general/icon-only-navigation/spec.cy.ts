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

  it('UI-FOUNDATION-SHELL-NAVIGATION-018/#1415 defaults primary navigation to a compact icon rail', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get('[data-slot="sidebar"]')
      .first()
      .should('have.attr', 'data-state', 'collapsed')

    const navLinks = [
      ['dashboard', 'Dashboard'],
      ['inventory', 'Inventory'],
      ['media', 'Media'],
      ['collections', 'Collections'],
      ['wishlist', 'Wishlist'],
      ['discoveries', 'Discoveries'],
      ['market-watch', 'Market Watch'],
      ['inbox', 'Inbox'],
      ['purchases', 'Purchases'],
      ['integrations', 'Integrations'],
      ['chats', 'Chats'],
      ['users', 'Users'],
      ['reports', 'Reports'],
    ] as const

    navLinks.forEach(([key, label]) => {
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

  it('UI-FOUNDATION-SHELL-NAVIGATION-019 keeps labels available after explicit expansion', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    cy.get('[data-slot="sidebar-trigger"]').first().click()
    cy.get('[data-slot="sidebar"]')
      .first()
      .should('have.attr', 'data-state', 'expanded')

    const navLinks = [
      ['dashboard', 'Dashboard'],
      ['inventory', 'Inventory'],
      ['media', 'Media'],
      ['collections', 'Collections'],
      ['wishlist', 'Wishlist'],
      ['discoveries', 'Discoveries'],
      ['market-watch', 'Market Watch'],
      ['inbox', 'Inbox'],
      ['purchases', 'Purchases'],
      ['integrations', 'Integrations'],
      ['chats', 'Chats'],
      ['users', 'Users'],
      ['reports', 'Reports'],
    ] as const

    navLinks.forEach(([key, label]) => {
      cy.get(`[data-testid="sidebar-nav-link-${key}"]`)
        .scrollIntoView()
        .should('be.visible')
        .and('have.attr', 'aria-label', label)
        .within(() => {
          cy.get('svg').should('be.visible')
          cy.get(`[data-testid="sidebar-nav-label-${key}"]`)
            .should('be.visible')
            .and('have.text', label)
        })
    })
  })
})
