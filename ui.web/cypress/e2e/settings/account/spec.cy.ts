describe('settings/account', () => {
  function signInToSettings() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/account',
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-ACCOUNT-001 persists account fields across reload', () => {
    cy.contains('button', 'Update account').should('not.be.disabled')
    cy.get('input[name="name"]').clear().type('Settings Account Name')
    cy.get('[data-testid="settings-account-language-trigger"]').click()
    cy.contains('[role="option"]', 'Japanese').click()
    cy.contains('button', 'Update account').click()
    cy.contains('Account settings saved.').should('be.visible')

    cy.reload()
    cy.get('input[name="name"]').should('have.value', 'Settings Account Name')
    cy.contains('button[role="combobox"]', 'Japanese').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-ACCOUNT-002 blocks invalid account submission with field errors', () => {
    cy.contains('button', 'Update account').should('not.be.disabled')
    cy.get('input[name="name"]').clear()
    cy.contains('button', 'Update account').click()
    cy.contains('Please enter your name.').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-ACCOUNT-004 retries account settings load failure without route reload', () => {
    let settingsAttempt = 0
    cy.intercept('GET', '/api/profiles/*/settings', (req) => {
      settingsAttempt += 1
      if (settingsAttempt === 1) {
        req.reply({
          statusCode: 503,
          body: { error: 'profile_settings_unavailable' },
        })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          settings: {
            'account.name': 'Retry Account Name',
            'account.language': 'ko',
            'account.dob': '1998-05-01T00:00:00.000Z',
          },
        },
      })
    }).as('accountSettings')

    cy.reload()
    cy.wait('@accountSettings')
    cy.contains('Failed to load account settings.').should('be.visible')
    cy.contains('button', 'Retry').click()
    cy.wait('@accountSettings')

    cy.location('pathname').should('match', /^\/settings\/account\/?$/)
    cy.contains('Failed to load account settings.').should('not.exist')
    cy.get('input[name="name"]').should('have.value', 'Retry Account Name')
    cy.contains('button[role="combobox"]', 'Korean').should('be.visible')
  })
})
