describe('general/profile-switcher-create-profile', () => {
  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-profile-create@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
  }

  it('PROFILES-004 creates and activates a new database profile from the switcher', () => {
    cy.request('POST', '/api/test/reset', {})
    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      expect(primaryResp.status).to.eq(201)
      cy.request('PUT', '/api/profiles/active', {
        profile_id: primaryResp.body.id,
      }).its('status').should('eq', 200)
    })

    cy.on('window:before:load', (win) => {
      cy.stub(win, 'prompt').returns('Travel DB')
    })

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get('[data-testid="active-profile-name"]').should('contain', 'Primary DB')

    cy.intercept('POST', '/api/profiles').as('createProfile')
    cy.intercept('PUT', '/api/profiles/active').as('activateProfile')

    cy.get('[data-testid="team-switcher-trigger"]').click()
    cy.get('[data-testid="team-switcher-add-profile"]').click()

    cy.wait('@createProfile').its('request.body').should('deep.equal', {
      name: 'Travel DB',
    })
    cy.wait('@activateProfile').its('request.body.profile_id').should('be.a', 'string')
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
