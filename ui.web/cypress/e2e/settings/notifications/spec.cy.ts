describe('settings/notifications', () => {
  function signInToSettings() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fnotifications')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/notifications\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-NOTIFICATIONS-001 persists editable notification controls', () => {
    cy.contains('button', 'Update notifications').should('not.be.disabled')
    cy.contains('label', 'Nothing').click()
    cy.get('[data-testid="settings-notifications-marketing"]').click()
    cy.contains('button', 'Update notifications').click()
    cy.contains('Notification settings saved.').should('be.visible')

    cy.request('/api/profiles/active').then((activeResp) => {
      const profileID = (activeResp.body?.id as string | undefined)?.trim() ?? ''
      expect(profileID).to.not.equal('')
      cy.request(`/api/profiles/${profileID}/settings`).then((settingsResp) => {
        expect(settingsResp.body.settings['notifications.type']).to.equal('none')
        expect(settingsResp.body.settings['notifications.marketing_emails']).to.equal(
          'true'
        )
      })
    })
  })

  it('UI-SCREEN-SETTINGS-NOTIFICATIONS-002 keeps guarded security control immutable', () => {
    cy.get('[data-testid="settings-notifications-security"]')
      .should('be.disabled')
      .and('have.attr', 'data-state', 'checked')
  })
})
