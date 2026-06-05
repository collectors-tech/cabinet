describe('settings/display', () => {
  function signInToSettings() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/display',
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-001 persists sidebar visibility selection', () => {
    cy.contains('button', 'Update display').should('not.be.disabled')
    cy.get('[data-testid="settings-display-desktop"]').click()
    cy.contains('button', 'Update display').click()
    cy.contains('Display settings saved.').should('be.visible')

    cy.reload()
    cy.get('[data-testid="settings-display-desktop"]')
      .should('have.attr', 'data-state', 'checked')
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-002 requires at least one selected display item', () => {
    cy.contains('button', 'Clear selection').click()
    cy.get('[data-testid^="settings-display-"][data-state="checked"]').should(
      'have.length',
      0
    )
    cy.contains('button', 'Update display').click()
    cy.contains('You have to select at least one item.').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-003 retries display settings load failure without route reload', () => {
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
            'display.items': 'home,documents',
          },
        },
      })
    }).as('displaySettings')

    cy.reload()
    cy.wait('@displaySettings')
    cy.contains('Failed to load display settings.').should('be.visible')
    cy.contains('button', 'Retry').click()
    cy.wait('@displaySettings')

    cy.location('pathname').should('match', /^\/settings\/display\/?$/)
    cy.contains('Failed to load display settings.').should('not.exist')
    cy.contains('button', 'Update display').should('not.be.disabled')
    cy.get('[data-testid="settings-display-home"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
    cy.get('[data-testid="settings-display-documents"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-004 clears selection without route transition', () => {
    cy.get('[data-testid^="settings-display-"][data-state="checked"]').should(
      'have.length.greaterThan',
      0
    )
    cy.contains('button', 'Clear selection').should('not.be.disabled').click()

    cy.location('pathname').should('match', /^\/settings\/display\/?$/)
    cy.get('[data-testid^="settings-display-"][data-state="checked"]').should(
      'have.length',
      0
    )
    cy.contains('button', 'Update display').should('not.be.disabled')
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-005 updates display with deterministic success feedback', () => {
    cy.intercept('PUT', '/api/profiles/*/settings').as('saveDisplay')

    cy.contains('button', 'Clear selection').click()
    cy.get('[data-testid="settings-display-documents"]').click()
    cy.get('[data-testid="settings-display-documents"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
    cy.contains('button', 'Update display').click()

    cy.wait('@saveDisplay')
      .its('request.body.settings')
      .should('deep.include', {
        'display.items': 'documents',
      })
    cy.contains('Display settings saved.').should('be.visible')

    cy.request('/api/profiles/active').then((activeResp) => {
      const profileID = (activeResp.body?.id as string | undefined)?.trim() ?? ''
      expect(profileID).to.not.equal('')
      cy.request(`/api/profiles/${profileID}/settings`).then((settingsResp) => {
        expect(settingsResp.body.settings['display.items']).to.equal(
          'documents'
        )
      })
    })
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-006 preserves selected items on save failure', () => {
    cy.intercept('PUT', '/api/profiles/*/settings', {
      statusCode: 500,
      body: { error: 'failed_to_update_settings' },
    }).as('displaySaveFailure')

    cy.contains('button', 'Clear selection').click()
    cy.get('[data-testid="settings-display-documents"]').click()
    cy.get('[data-testid="settings-display-downloads"]').click()
    cy.contains('button', 'Update display').click()

    cy.wait('@displaySaveFailure')
    cy.contains('profile_settings_save_500').should('be.visible')
    cy.get('[data-testid="settings-display-documents"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
    cy.get('[data-testid="settings-display-downloads"]').should(
      'have.attr',
      'data-state',
      'checked'
    )
    cy.get('[data-testid="settings-display-home"]').should(
      'have.attr',
      'data-state',
      'unchecked'
    )
  })
})
