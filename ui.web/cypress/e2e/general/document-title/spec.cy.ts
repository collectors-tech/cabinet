describe('document title', () => {
  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-document-title@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
  }

  it('uses Cabinet - <Page Title> across primary authenticated routes', () => {
    const primaryRoutes = [
      { path: '/dashboard/', title: 'Cabinet - Home' },
      { path: '/inventory/', title: 'Cabinet - Inventory' },
      { path: '/collections/', title: 'Cabinet - Collections' },
      { path: '/wishlist/', title: 'Cabinet - Wishlist' },
      { path: '/integrations/', title: 'Cabinet - Integrations' },
      { path: '/chats/', title: 'Cabinet - Chats' },
      { path: '/inbox/', title: 'Cabinet - Inbox' },
      { path: '/discoveries/', title: 'Cabinet - Discoveries' },
      { path: '/reports/', title: 'Cabinet - Reports' },
      { path: '/settings/profile', title: 'Cabinet - Settings' },
    ]

    cy.clearCookies()
    cy.clearLocalStorage()
    signInTo(primaryRoutes[0].path)
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/dashboard\/?$/)

    primaryRoutes.forEach(({ path, title }) => {
      cy.visit(path)
      cy.location('pathname', { timeout: 15000 }).should('eq', path)
      cy.title().should('eq', title)
      cy.title().should('not.match', /^[a-z]+\./)
    })
  })
})
