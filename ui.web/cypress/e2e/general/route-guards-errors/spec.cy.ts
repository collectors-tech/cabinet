describe('route guards, titles, and error pages', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
  })

  function stubLocalSession(profileId: string) {
    expect(profileId).to.not.equal('')
    cy.intercept('POST', '**/api/auth/local/session', {
      statusCode: 200,
      body: {
        ok: true,
        session_token:
          'test-only-opaque-profile-bound-session-credential-000000000001',
      },
    }).as('routeGuardLocalSession')
  }

  function signInTo(path: string) {
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.e2eSetSetupState('present')
      cy.e2eEnsureSignedOut()
      stubLocalSession(profile_id)
      cy.useBootstrappedProfile(profile_id, profile_name, { path })
    })
  }

  it('UI-ROUTE-COVERAGE-004 preserves protected-route deep-link query state through the sign-in guard', () => {
    cy.e2eBootstrap().then(({ profile_id }) => {
      cy.e2eEnsureSignedOut()
      stubLocalSession(profile_id)

      cy.visit('/inventory?view=table&focus=filters')

      cy.location('pathname', { timeout: 15000 }).should('eq', '/sign-in')
      cy.location('search').should('include', 'redirect=')
      cy.location('search').should(
        'include',
        encodeURIComponent('/inventory?view=table&focus=filters')
      )
      cy.get('[data-testid="local-device-auth-boundary"]').should('be.visible')
      cy.get('input[name="email"]').should('not.exist')
      cy.get('input[name="password"]').should('not.exist')
      cy.contains('button', 'Open local workspace').should('be.visible').click()
      cy.location('pathname', { timeout: 15000 }).should(
        'match',
        /^\/inventory\/?$/
      )
      cy.location('search').should('eq', '?view=table&focus=filters')
    })
  })

  it('UI-ROUTE-COVERAGE-005 shows public 404 recovery controls', () => {
    cy.e2eEnsureSignedOut()

    cy.visit('/route-that-does-not-exist', { failOnStatusCode: false })

    cy.contains('h1', '404').should('be.visible')
    cy.contains('Oops! Page Not Found!').should('be.visible')
    cy.contains('button', 'Go Back').should('be.visible')
    cy.contains('button', 'Back to Home').should('be.visible').click()
    cy.location('pathname', { timeout: 15000 }).should('eq', '/sign-in')
  })

  it('UI-ROUTE-COVERAGE-005 renders authenticated error taxonomy states with shell recovery', () => {
    signInTo('/dashboard')

    cy.visit('/errors/forbidden?e2e=route-guards-errors')
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/errors\/forbidden\/?$/
    )
    cy.get('[data-testid="error-header-title"]')
      .should('be.visible')
      .and('contain', 'Error')
    cy.contains('h1', '403').should('be.visible')
    cy.contains('Access Forbidden').should('be.visible')
    cy.contains('button', 'Go Back').should('be.visible')
    cy.contains('button', 'Back to Home').should('be.visible').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/dashboard\/?$/)

    cy.visit('/errors/maintenance-error?e2e=route-guards-errors')
    cy.contains('h1', '503').should('be.visible')
    cy.contains('Website is under maintenance!').should('be.visible')
    cy.contains('button', 'Learn more').should('be.visible')
  })

  it('UI-ROUTE-COVERAGE-006 keeps Cabinet-prefixed document titles across route classes', () => {
    signInTo('/dashboard')

    const routeTitles = [
      { path: '/dashboard', title: 'Cabinet - Home' },
      { path: '/settings/storage', title: 'Cabinet - Storage Settings' },
      {
        path: '/errors/not-found?e2e=route-guards-errors',
        title: 'Cabinet - Error',
      },
      { path: '/404?e2e=route-guards-errors', title: 'Cabinet - Not Found' },
    ]

    routeTitles.forEach(({ path, title }) => {
      cy.visit(path)
      cy.title().should('eq', title)
      cy.title().should('not.match', /^[a-z]+\./)
    })
  })
})
