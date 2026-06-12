describe('settings/appearance', () => {
  function signInToSettings() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/appearance',
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-004 defaults first-run theme to dark when no preference exists', () => {
    cy.clearCookie('vite-ui-theme')
    cy.clearLocalStorage()
    cy.visit('/sign-in', {
      onBeforeLoad(win) {
        Object.defineProperty(win, 'matchMedia', {
          writable: true,
          value: (query: string) => ({
            matches: false,
            media: query,
            onchange: null,
            addListener: () => {},
            removeListener: () => {},
            addEventListener: () => {},
            removeEventListener: () => {},
            dispatchEvent: () => false,
          }),
        })
      },
    })
    cy.getCookie('vite-ui-theme').should('not.exist')
    cy.get('html').should('have.class', 'dark')
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-001 applies and persists theme/font preferences', () => {
    signInToSettings()
    cy.contains('button', 'Update preferences').should('not.be.disabled')
    cy.get('select[name="font"]').select('manrope')
    cy.contains('span', 'Light').click()
    cy.contains('button', 'Update preferences').click()

    cy.get('html').should('have.class', 'light')
    cy.reload()
    cy.get('select[name="font"]').should('have.value', 'manrope')
    cy.get('html').should('have.class', 'light')
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-002 supports Chinese and Japanese selection with persistence', () => {
    signInToSettings()

    cy.get('[data-testid="appearance-language-select"]').select('zh')
    cy.contains('button', 'Update preferences').click()
    cy.window().its('localStorage.i18nextLng').should('eq', 'zh')

    cy.get('[data-testid="appearance-language-select"]').select('ja')
    cy.contains('button', 'Update preferences').click()
    cy.window().its('localStorage.i18nextLng').should('eq', 'ja')

    cy.reload()
    cy.get('[data-testid="appearance-language-select"]').should('have.value', 'ja')
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-003 falls back to English text when key is missing in selected locale', () => {
    signInToSettings()

    cy.get('[data-testid="appearance-language-select"]').select('zh')
    cy.contains('button', 'Update preferences').click()
    cy.get('[data-testid="appearance-language-fallback-sample"]').should(
      'contain',
      'Fallback sample text from English defaults.'
    )
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-005 updates preferences with deterministic success feedback', () => {
    signInToSettings()
    cy.intercept('PUT', '/api/profiles/*/settings').as('saveAppearance')

    cy.contains('button', 'Update preferences').should('not.be.disabled')
    cy.get('select[name="font"]').select('inter')
    cy.get('[data-testid="appearance-language-select"]').select('ja')
    cy.contains('span', 'Light').click()
    cy.contains('button', 'Update preferences').click()

    cy.wait('@saveAppearance')
      .its('request.body.settings')
      .should('deep.include', {
        'appearance.theme': 'light',
        'appearance.font': 'inter',
        'appearance.language': 'ja',
      })
    cy.contains('Appearance settings saved.').should('be.visible')
    cy.get('html').should('have.class', 'light')
    cy.window().its('localStorage.i18nextLng').should('eq', 'ja')
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-005 retries appearance settings load failure without route reload', () => {
    signInToSettings()
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
            'appearance.theme': 'light',
            'appearance.font': 'manrope',
            'appearance.language': 'zh',
          },
        },
      })
    }).as('appearanceSettings')

    cy.reload()
    cy.wait('@appearanceSettings')
    cy.contains('Failed to load appearance settings.').should('be.visible')
    cy.contains('button', 'Retry').click()
    cy.wait('@appearanceSettings')

    cy.location('pathname').should('match', /^\/settings\/appearance\/?$/)
    cy.contains('Failed to load appearance settings.').should('not.exist')
    cy.get('select[name="font"]').should('have.value', 'manrope')
    cy.get('[data-testid="appearance-language-select"]').should('have.value', 'zh')
    cy.contains('span', 'Light').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-006 preserves edited appearance controls without applying unpersisted preferences when save fails', () => {
    signInToSettings()
    cy.intercept('PUT', '/api/profiles/*/settings', {
      statusCode: 503,
      body: { error: 'appearance_settings_save_unavailable' },
    }).as('saveAppearanceFailure')

    cy.get('html').should('have.class', 'dark')
    cy.window().its('localStorage.i18nextLng').should('not.eq', 'ja')
    cy.get('select[name="font"]').should('not.have.value', 'manrope')

    cy.get('select[name="font"]').select('manrope')
    cy.get('[data-testid="appearance-language-select"]').select('ja')
    cy.contains('span', 'Light').click()
    cy.contains('button', 'Update preferences').click()

    cy.wait('@saveAppearanceFailure')
      .its('request.body.settings')
      .should('deep.include', {
        'appearance.theme': 'light',
        'appearance.font': 'manrope',
        'appearance.language': 'ja',
      })
    cy.contains('profile_settings_save_503').should('be.visible')
    cy.contains('Appearance settings saved.').should('not.exist')
    cy.get('select[name="font"]').should('have.value', 'manrope')
    cy.get('[data-testid="appearance-language-select"]').should('have.value', 'ja')
    cy.get('html').should('have.class', 'dark')
    cy.window().its('localStorage.i18nextLng').should('not.eq', 'ja')
    cy.contains('button', 'Update preferences').should('not.be.disabled')
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-007 blocks appearance edits when active profile is missing', () => {
    signInToSettings()
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 404,
      body: { error: 'active_profile_404' },
    }).as('activeProfileMissing')

    cy.reload()
    cy.wait('@activeProfileMissing')

    cy.location('pathname').should('match', /^\/settings\/appearance\/?$/)
    cy.get('[data-testid="settings-profile-context-blocked"]').should(
      'be.visible'
    )
    cy.contains('Active profile is required.').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
    cy.contains('a', 'Create or Select Profile').should('be.visible')
    cy.contains('button', 'Update preferences').should('not.exist')
    cy.contains('label', 'Font').should('not.exist')
    cy.get('[data-testid="appearance-language-select"]').should('not.exist')
    cy.contains('span', 'Light').should('not.exist')

    cy.contains('button', 'Retry').click()
    cy.wait('@activeProfileMissing')
  })
})
