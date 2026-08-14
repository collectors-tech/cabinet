describe('general/profile-switcher-existing-profile-context', () => {
  function enterWithProfile(path: string, profileId: string, profileName: string) {
    cy.e2eEnsureSignedOut()
    cy.intercept('POST', '/api/auth/local/session', (request) => {
      expect(request.body.profile_id).to.be.a('string')
      request.reply({
        statusCode: 200,
        body: {
          ok: true,
          session_token:
            'test-only-opaque-profile-bound-session-credential-000000000001',
        },
      })
    }).as('localServerSession')
    cy.useBootstrappedProfile(profileId, profileName, { path })
  }

  function expectSelectedProfileOn(
    path: string,
    expectedName: string,
    expectedStatus = 'Database'
  ) {
    cy.visit(path)
    cy.get('[data-testid="active-profile-name"]', { timeout: 15000 }).should(
      'contain',
      expectedName
    )
    cy.get('[data-testid="active-profile-status"]').should(
      'contain',
      expectedStatus
    )
    cy.request('GET', '/api/profiles/active')
      .its('body.name')
      .should('eq', expectedName)
  }

  it('PROFILES-005 selects an existing database profile across app sections', () => {
    let primaryID = ''
    cy.request('POST', '/api/test/reset', {})
    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      expect(primaryResp.status).to.eq(201)
      primaryID = primaryResp.body.id as string

      cy.request('POST', '/api/profiles', { name: 'Showcase DB' }).then((showcaseResp) => {
        expect(showcaseResp.status).to.eq(201)
        cy.request('PUT', '/api/profiles/active', { profile_id: primaryID })
          .its('status')
          .should('eq', 200)
      })
    })

    cy.then(() => {
      enterWithProfile('/inventory/', primaryID, 'Primary DB')
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get('[data-testid="active-profile-name"]').should('contain', 'Primary DB')

    cy.intercept('PUT', '/api/profiles/active').as('activateProfile')
    cy.get('[data-testid="team-switcher-trigger"]').click()
    cy.get('[data-testid="team-option-primary-db-plan"]').should('contain', 'Database')
    cy.get('[data-testid="team-option-showcase-db-plan"]').should(
      'contain',
      'Showcase sample data'
    )
    cy.get('[data-testid="team-option-showcase-db"]').click()
    cy.wait('@activateProfile').its('request.body.profile_id').should('be.a', 'string')
    cy.get('[data-testid="active-profile-name"]', { timeout: 20000 }).should(
      'contain',
      'Showcase DB'
    )
    cy.get('[data-testid="active-profile-status"]').should(
      'contain',
      'Showcase sample data'
    )

    cy.request('GET', '/api/profiles/active')
      .its('body.name')
      .should('eq', 'Showcase DB')

    ;[
      '/inventory/',
      '/wishlist/',
      '/collections/',
      '/settings/profile',
      '/chats/',
      '/integrations/',
    ].forEach((path) => {
      expectSelectedProfileOn(path, 'Showcase DB', 'Showcase sample data')
    })
  })
})
