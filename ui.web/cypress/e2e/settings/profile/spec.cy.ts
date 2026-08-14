describe('settings/profile', () => {
  function signInToSettings() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/profile',
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-PROFILE-001 persists profile values through Cabinet settings API', () => {
    cy.contains('Profile settings').should('be.visible')
    cy.contains(
      'Manage your account profile and public display details.'
    ).should('be.visible')
    cy.contains('settings.profile.title').should('not.exist')
    cy.contains('settings.profile.description').should('not.exist')
    cy.contains('button', 'Update profile').should('not.be.disabled')
    cy.get('input[name="username"]').clear().type('collector-profile')
    cy.get('[data-testid="settings-profile-email-trigger"]').click()
    cy.get('[role="option"]').contains('m@support.com').click()
    cy.get('textarea[name="bio"]')
      .clear()
      .type('Collector profile bio for e2e.')
    cy.contains('button', 'Add URL').click()
    cy.get('input[name="urls.0.value"]').type(
      'https://collector.example/profile'
    )
    cy.get('[data-testid="settings-profile-telegram-capture"]').should(
      'be.visible'
    )
    cy.get('[data-testid="settings-profile-telegram-capture"]')
      .should('contain.text', 'private-chat pairing')
      .and('contain.text', 'instead of manually entered sender or chat IDs')
    cy.get('[data-testid="settings-profile-telegram-sender-id"]').should(
      'not.exist'
    )
    cy.get('[data-testid="settings-profile-telegram-chat-id"]').should(
      'not.exist'
    )
    cy.get('[data-testid="settings-profile-open-telegram-setup"]')
      .should('have.attr', 'href')
      .and('eq', '/integrations?filter=telegram')
    cy.contains('button', 'Update profile').click()
    cy.contains('Profile settings saved.').should('be.visible')

    cy.reload()
    cy.get('input[name="username"]').should('have.value', 'collector-profile')
    cy.get('textarea[name="bio"]').should(
      'have.value',
      'Collector profile bio for e2e.'
    )
    cy.get('[data-testid="settings-profile-email-trigger"]').should(
      'contain.text',
      'm@support.com'
    )
    cy.get('input[name="urls.0.value"]').should(
      'have.value',
      'https://collector.example/profile'
    )
    cy.get('[data-testid="settings-profile-telegram-sender-id"]').should(
      'not.exist'
    )
    cy.get('[data-testid="settings-profile-telegram-chat-id"]').should(
      'not.exist'
    )
  })

  it('UI-SCREEN-SETTINGS-PROFILE-002 preserves entered values on profile save failure', () => {
    cy.intercept('PUT', '/api/profiles/*/settings', {
      statusCode: 500,
      body: { error: 'failed_to_update_settings' },
    }).as('profileSaveFailure')

    cy.contains('button', 'Update profile').should('not.be.disabled')
    cy.get('input[name="username"]').clear().type('failed-profile-save')
    cy.get('textarea[name="bio"]')
      .clear()
      .type('This value must remain after failure.')
    cy.contains('button', 'Update profile').click()
    cy.wait('@profileSaveFailure')
    cy.contains('profile_settings_save_500').should('be.visible')
    cy.get('input[name="username"]').should('have.value', 'failed-profile-save')
    cy.get('textarea[name="bio"]').should(
      'have.value',
      'This value must remain after failure.'
    )
  })

  it('UI-SCREEN-SETTINGS-PROFILE-001 blocks invalid profile URL submission before save', () => {
    let profileSaveCalls = 0
    cy.intercept('PUT', '/api/profiles/*/settings', (req) => {
      profileSaveCalls += 1
      req.reply({
        statusCode: 500,
        body: { error: 'unexpected_profile_save' },
      })
    })

    cy.contains('button', 'Update profile').should('not.be.disabled')
    cy.get('input[name="username"]').clear().type('collector-profile')
    cy.get('textarea[name="bio"]')
      .clear()
      .type('Collector profile bio for invalid URL coverage.')
    cy.contains('button', 'Add URL').click()
    cy.get('input[name="urls.0.value"]').type('not-a-url')
    cy.contains('button', 'Update profile').click()

    cy.contains('Please enter a valid URL.').should('be.visible')
    cy.get('input[name="urls.0.value"]').should('have.value', 'not-a-url')
    cy.then(() => {
      expect(profileSaveCalls).to.equal(0)
    })
  })

  it('UI-SCREEN-SETTINGS-PROFILE-003 retries profile settings load failure without route reload', () => {
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
            'profile.username': 'retry-profile',
            'profile.email': 'm@example.com',
            'profile.bio': 'Retry recovered profile settings.',
            'profile.urls': JSON.stringify([
              { value: 'https://collector.example/recovered' },
            ]),
          },
        },
      })
    }).as('profileSettings')

    cy.reload()
    cy.wait('@profileSettings')
    cy.contains('Failed to load profile settings.').should('be.visible')
    cy.contains('button', 'Retry').click()
    cy.wait('@profileSettings')

    cy.location('pathname').should('match', /^\/settings\/profile\/?$/)
    cy.contains('Failed to load profile settings.').should('not.exist')
    cy.get('input[name="username"]').should('have.value', 'retry-profile')
    cy.get('textarea[name="bio"]').should(
      'have.value',
      'Retry recovered profile settings.'
    )
    cy.get('input[name="urls.0.value"]').should(
      'have.value',
      'https://collector.example/recovered'
    )
  })

  it('UI-SCREEN-SETTINGS-PROFILE-005 blocks profile edits when active profile is missing', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 404,
      body: { error: 'active_profile_404' },
    }).as('activeProfileMissing')

    cy.visit('/settings/profile')
    cy.wait('@activeProfileMissing')

    cy.location('pathname').should('match', /^\/settings\/profile\/?$/)
    cy.get('[data-testid="settings-profile-context-blocked"]').should(
      'be.visible'
    )
    cy.contains('Active profile is required.').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
    cy.contains('a', 'Create or Select Profile').should('be.visible')
    cy.get('input[name="username"]').should('not.exist')
    cy.get('[data-testid="settings-profile-email-trigger"]').should('not.exist')
    cy.get('textarea[name="bio"]').should('not.exist')
    cy.get('[data-testid="settings-profile-telegram-capture"]').should(
      'not.exist'
    )
    cy.contains('button', 'Add URL').should('not.exist')
    cy.contains('button', 'Update profile').should('not.exist')

    cy.contains('button', 'Retry').click()
    cy.wait('@activeProfileMissing')
  })
})
