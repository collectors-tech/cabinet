describe('settings/notifications', () => {
  function signInToSettings() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/notifications',
    })
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

  it('UI-SCREEN-SETTINGS-NOTIFICATIONS-003 updates notifications with deterministic success feedback', () => {
    cy.intercept('PUT', '/api/profiles/*/settings').as('saveNotifications')

    cy.contains('button', 'Update notifications').should('not.be.disabled')
    cy.contains('label', 'All new messages').click()
    cy.get('[data-testid="settings-notifications-communication"]').click()
    cy.get('[data-testid="settings-notifications-social"]').click()
    cy.contains('button', 'Update notifications').click()

    cy.wait('@saveNotifications')
      .its('request.body.settings')
      .should('deep.include', {
        'notifications.type': 'all',
        'notifications.communication_emails': 'true',
        'notifications.social_emails': 'true',
        'notifications.security_emails': 'true',
      })
    cy.contains('Notification settings saved.').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-NOTIFICATIONS-003 retries notifications settings load failure without route reload', () => {
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
            'notifications.type': 'none',
            'notifications.mobile': 'true',
            'notifications.communication_emails': 'true',
            'notifications.social_emails': 'false',
            'notifications.marketing_emails': 'true',
            'notifications.security_emails': 'true',
          },
        },
      })
    }).as('notificationSettings')

    cy.reload()
    cy.wait('@notificationSettings')
    cy.contains('Failed to load notification settings.').should('be.visible')
    cy.contains('button', 'Retry').click()
    cy.wait('@notificationSettings')

    cy.location('pathname').should('match', /^\/settings\/notifications\/?$/)
    cy.contains('Failed to load notification settings.').should('not.exist')
    cy.contains('button', 'Update notifications').should('not.be.disabled')
    cy.contains('label', 'Nothing').click()
    cy.get('[data-testid="settings-notifications-communication"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
    cy.get('[data-testid="settings-notifications-marketing"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
  })
})
