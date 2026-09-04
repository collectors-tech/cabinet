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
    cy.window()
      .its('localStorage')
      .invoke('getItem', 'cabinet.toastHistory.v1')
      .then((rawHistory) => {
        expect(rawHistory).to.be.a('string')
        const history = JSON.parse(rawHistory as string) as Array<{
          title?: string
          level?: string
          source_label?: string
          category?: string
          summary?: string
        }>
        const savedNotification = history.find(
          (record) => record.title === 'Notification settings saved.'
        )
        expect(savedNotification).to.deep.include({
          title: 'Notification settings saved.',
          level: 'success',
          source_label: 'Settings Notifications',
          category: 'settings',
          summary:
            'Settings Notifications preferences were persisted and preserved in Inbox history.',
        })
      })
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

  it('UI-SCREEN-SETTINGS-NOTIFICATIONS-004 preserves edited notification choices when save fails', () => {
    const setSwitchState = (testId: string, checked: boolean) => {
      cy.get(`[data-testid="${testId}"]`).then(($switch) => {
        if (($switch.attr('data-state') === 'checked') !== checked) {
          cy.wrap($switch).click()
        }
      })
    }

    cy.intercept('PUT', '/api/profiles/*/settings', {
      statusCode: 503,
      body: { error: 'notification_settings_save_unavailable' },
    }).as('saveNotificationsFailure')

    cy.contains('button', 'Update notifications').should('not.be.disabled')
    cy.contains('label', 'Nothing').click()
    setSwitchState('settings-notifications-communication', true)
    setSwitchState('settings-notifications-marketing', true)
    setSwitchState('settings-notifications-social', false)
    cy.get('[data-testid="settings-notifications-mobile"]').click()
    cy.contains('button', 'Update notifications').click()

    cy.wait('@saveNotificationsFailure')
      .its('request.body.settings')
      .should('deep.include', {
        'notifications.type': 'none',
        'notifications.mobile': 'true',
        'notifications.communication_emails': 'true',
        'notifications.marketing_emails': 'true',
        'notifications.social_emails': 'false',
        'notifications.security_emails': 'true',
      })
    cy.contains('profile_settings_save_503').should('be.visible')
    cy.contains('Notification settings saved.').should('not.exist')
    cy.location('pathname').should('match', /^\/settings\/notifications\/?$/)
    cy.get('input[value="none"]').should('be.checked')
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
    cy.get('[data-testid="settings-notifications-social"]').should(
      'have.attr',
      'data-state',
      'unchecked'
    )
    cy.get('[data-testid="settings-notifications-mobile"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
    cy.contains('button', 'Update notifications').should('not.be.disabled')
  })

  it('UI-SCREEN-SETTINGS-NOTIFICATIONS-005 blocks notification edits when active profile is missing', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 404,
      body: { error: 'active_profile_404' },
    }).as('activeProfileMissing')

    cy.visit('/settings/notifications')
    cy.wait('@activeProfileMissing')

    cy.location('pathname').should('match', /^\/settings\/notifications\/?$/)
    cy.get('[data-testid="settings-profile-context-blocked"]').should(
      'be.visible'
    )
    cy.contains('Active profile is required.').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
    cy.contains('a', 'Create or Select Profile').should('be.visible')
    cy.contains('button', 'Update notifications').should('not.exist')
    cy.contains('label', 'Notify me about...').should('not.exist')
    cy.get('[data-testid="settings-notifications-communication"]').should(
      'not.exist'
    )
    cy.get('[data-testid="settings-notifications-mobile"]').should('not.exist')

    cy.contains('button', 'Retry').click()
    cy.wait('@activeProfileMissing')
  })
})
