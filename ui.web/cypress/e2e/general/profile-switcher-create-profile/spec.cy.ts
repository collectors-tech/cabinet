describe('general/profile-switcher-create-profile', () => {
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

  it('PROFILES-004 creates and activates a new database profile from the switcher', () => {
    let primaryID = ''
    cy.request('POST', '/api/test/reset', {})
    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      expect(primaryResp.status).to.eq(201)
      primaryID = primaryResp.body.id as string
      cy.request('PUT', '/api/profiles/active', {
        profile_id: primaryID,
      }).its('status').should('eq', 200)
    })

    cy.on('window:before:load', (win) => {
      cy.stub(win, 'prompt').returns('Travel DB')
    })

    cy.then(() => {
      enterWithProfile('/inventory/', primaryID, 'Primary DB')
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get('[data-testid="active-profile-name"]').should('contain', 'Primary DB')

    cy.intercept({ method: 'POST', pathname: '/api/profiles' }).as('createProfile')

    cy.get('[data-testid="team-switcher-trigger"]').click()
    cy.get('[data-testid="team-switcher-add-profile"]').click()

    cy.wait('@createProfile').its('request.body').should('deep.equal', {
      name: 'Travel DB',
    })
    cy.get('[data-testid="active-profile-name"]', { timeout: 20000 }).should(
      'contain',
      'Travel DB'
    )

    cy.request('GET', '/api/profiles/active')
      .its('body')
      .should((active) => {
        expect(active.name).to.eq('Travel DB')
      })
    cy.request('GET', '/api/profiles')
      .its('body.profiles')
      .should((profiles) => {
        const names = profiles.map((profile: { name?: string }) => profile.name)
        expect(names).to.include('Travel DB')
      })
  })
})
