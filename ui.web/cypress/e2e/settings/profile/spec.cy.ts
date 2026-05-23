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
    cy.contains('Manage your account profile and public display details.').should(
      'be.visible'
    )
    cy.contains('settings.profile.title').should('not.exist')
    cy.contains('settings.profile.description').should('not.exist')
    cy.contains('button', 'Update profile').should('not.be.disabled')
    cy.get('input[name="username"]').clear().type('collector-profile')
    cy.get('[data-testid="settings-profile-email-trigger"]').click()
    cy.get('[role="option"]').contains('m@support.com').click()
    cy.get('textarea[name="bio"]')
      .clear()
      .type('Collector profile bio for e2e.')
    cy.get('[data-testid="settings-profile-telegram-capture"]').should(
      'be.visible'
    )
    cy.get('[data-testid="settings-profile-telegram-sender-id"]')
      .clear()
      .type('255192091')
    cy.get('[data-testid="settings-profile-telegram-chat-id"]')
      .clear()
      .type('-5235769556')
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
    cy.get('[data-testid="settings-profile-telegram-sender-id"]').should(
      'have.value',
      '255192091'
    )
    cy.get('[data-testid="settings-profile-telegram-chat-id"]').should(
      'have.value',
      '-5235769556'
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
})
