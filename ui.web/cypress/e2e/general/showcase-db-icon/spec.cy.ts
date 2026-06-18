describe('showcase-db-icon', () => {
  function visibleByTestId(testId: string) {
    return cy.get(`[data-testid="${testId}"]`).first()
  }

  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-showcase-db-icon@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.request('POST', '/api/test/reset', {})
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-010 renders Showcase DB icon variants with accessible profile text', () => {
    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      expect(primaryResp.status).to.eq(201)

      cy.request('POST', '/api/profiles', { name: 'Showcase DB' }).then((showcaseResp) => {
        expect(showcaseResp.status).to.eq(201)
        const showcaseID = showcaseResp.body.id as string

        cy.request('PUT', '/api/profiles/active', { profile_id: showcaseID }).its('status').should('eq', 200)
      })
    })

    signInTo('/inventory/')
    visibleByTestId('active-profile-name').should('contain', 'Showcase DB')
    cy.get('[data-testid="active-profile-status"]').should(
      'contain',
      'Showcase sample data'
    )
    cy.get('[data-testid="active-profile-db-icon"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Showcase DB database profile')
    cy.get('[data-testid="active-profile-db-icon-variant"]').should(
      'have.attr',
      'data-db-icon-variant',
      'dark'
    )

    visibleByTestId('team-switcher-trigger').click()

    cy.get('[data-testid="team-option-showcase-db-icon"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Showcase DB database profile')
    cy.get('[data-testid="team-option-showcase-db-icon-light"]').should(
      'have.attr',
      'data-db-icon-variant',
      'light'
    )
    cy.get('[data-testid="team-option-showcase-db-icon-dark"]').should(
      'have.attr',
      'data-db-icon-variant',
      'dark'
    )
  })
})
