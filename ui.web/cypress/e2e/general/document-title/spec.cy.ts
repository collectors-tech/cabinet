describe('document title', () => {
  function signInTo(path: string) {
    cy.e2eReset()
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.e2eSetSetupState('present')
      cy.e2eEnsureSignedOut()
      cy.intercept('POST', '/api/auth/local/session', (request) => {
        expect(String(request.body?.profile_id ?? '')).to.match(/\S/)
        request.reply({
          statusCode: 200,
          body: {
            ok: true,
            session_token:
              'test-only-opaque-profile-bound-session-credential-000000000001',
          },
        })
      })
      cy.useBootstrappedProfile(profile_id, profile_name, { path })
    })
  }

  it('uses Cabinet - <Page Title> across primary authenticated routes', () => {
    const primaryRoutes = [
      { path: '/dashboard/', title: 'Cabinet - Home' },
      { path: '/inventory/', title: 'Cabinet - Inventory' },
      { path: '/media/', title: 'Cabinet - Media' },
      { path: '/collections/', title: 'Cabinet - Collections' },
      { path: '/wishlist/', title: 'Cabinet - Wishlist' },
      { path: '/integrations/', title: 'Cabinet - Integrations' },
      { path: '/chats/', title: 'Cabinet - Chats' },
      { path: '/inbox/', title: 'Cabinet - Notification Inbox' },
      { path: '/discoveries/', title: 'Cabinet - Discoveries' },
      { path: '/reports/', title: 'Cabinet - Reports' },
      { path: '/settings/profile', title: 'Cabinet - Profile Settings' },
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
