describe('showcase-db-icon', () => {
  function visibleByTestId(testId: string) {
    return cy.get(`[data-testid="${testId}"]`).first()
  }

  function enterShowcaseProfile() {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')

    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      expect(primaryResp.status).to.eq(201)

      cy.request('POST', '/api/profiles', { name: 'Showcase DB' }).then((showcaseResp) => {
        expect(showcaseResp.status).to.eq(201)
        const showcaseID = showcaseResp.body.id as string

        cy.request('PUT', '/api/profiles/active', {
          profile_id: primaryResp.body.id,
        }).its('status').should('eq', 200)
        cy.e2eEnsureSignedOut()
        cy.stubLocalServerSession(showcaseID)
        cy.setCookie('sidebar_state', 'true')
        cy.useBootstrappedProfile(showcaseID, 'Showcase DB', {
          path: '/inventory/',
        })
      })
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-010 renders Showcase DB icon variants with accessible profile text', () => {
    enterShowcaseProfile()
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
