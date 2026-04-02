describe('general/ui-foundation-components', () => {
  function bootstrapAndSignIn(path = '/settings/profile') {
    cy.e2eReset()
    cy.e2eBootstrap().then((bootstrap) => {
      cy.useBootstrappedProfile(bootstrap.profile_id, bootstrap.profile_name, { path })
    })
  }

  it('UI-FOUNDATION-COMPONENTS-001 exposes explicit foundation component contract surface on settings profile', () => {
    bootstrapAndSignIn('/settings/profile')
    cy.get('main').within(() => {
      cy.contains('h1', /^\s*Settings\s*$/).should('be.visible')
      cy.get('input[placeholder="cabinet-user"]').should('be.visible')
      cy.contains('button', 'Update profile').should('be.visible')
    })

    cy.get('[aria-label="Open theme settings"]').click()
    cy.contains('Theme Settings').should('be.visible')
    cy.get('body').type('{esc}')
    cy.contains('Theme Settings').should('not.exist')
  })

  it('UI-FOUNDATION-COMPONENTS-002 renders deterministic error and retry-ready behavior for data-bound components', () => {
    let usersAttempts = 0
    cy.intercept('GET', '/api/users*', (req) => {
      usersAttempts += 1
      if (usersAttempts === 1) {
        req.reply({ statusCode: 500, body: { error: 'users_failed' } })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          users: [
            {
              id: 'user-components-1',
              firstName: 'Retry',
              lastName: 'Ready',
              username: 'retry.ready',
              email: 'retry.ready@example.com',
              phoneCode: '+61',
              phoneNumber: '000000000',
              role: 'view',
              status: 'active',
            },
          ],
        },
      })
    }).as('usersLoad')

    bootstrapAndSignIn('/users/')
    cy.wait('@usersLoad')
    cy.get('[data-testid="users-load-error"]').should('be.visible')
    cy.contains('button', 'Retry').click()
    cy.wait('@usersLoad')
    cy.get('[data-testid="users-load-error"]').should('not.exist')
    cy.contains('Retry').should('be.visible')
  })

  it('UI-FOUNDATION-COMPONENTS-003 prevents duplicate submit while save request is in-flight', () => {
    const targetUsername = 'component-lock-user'
    cy.intercept('PUT', '/api/profiles/*/settings', (req) => {
      req.reply({ delay: 1200, statusCode: 200, body: req.body })
    }).as('saveSettings')

    bootstrapAndSignIn('/settings/profile')
    cy.get('input[placeholder="cabinet-user"]').clear().type(targetUsername)
    cy.contains('button', 'Update profile').dblclick()
    cy.contains('button', 'Update profile').should('be.disabled')
    cy.wait('@saveSettings')
    cy.contains('Profile settings saved.').should('be.visible')
  })

  it('UI-FOUNDATION-COMPONENTS-004 enforces keyboard dialog open/escape-close with trigger focus restore', () => {
    bootstrapAndSignIn('/settings/profile')
    cy.get('[aria-label="Open theme settings"]').as('themeTrigger').click()
    cy.contains('Theme Settings').should('be.visible')
    cy.focused().type('{esc}')
    cy.contains('Theme Settings').should('not.exist')
    cy.get('@themeTrigger').should('be.focused')
  })

  it('UI-FOUNDATION-COMPONENTS-005 links component contract testability artifacts to executable coverage', () => {
    bootstrapAndSignIn('/settings/profile')
    cy.contains('button', 'Update profile').should('be.visible')
    cy.get('[aria-label="Open theme settings"]').should('be.visible')
    cy.visit('/users')
    cy.contains('Users').should('be.visible')
  })
})
