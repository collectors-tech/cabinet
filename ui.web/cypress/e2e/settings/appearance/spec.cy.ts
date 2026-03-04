describe('settings/appearance', () => {
  function signInToSettings() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fappearance')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/appearance\/?$/
    )
  }

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
    cy.clearCookies()
    cy.clearLocalStorage()
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
    cy.clearCookies()
    cy.clearLocalStorage()
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
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()

    cy.get('[data-testid="appearance-language-select"]').select('zh')
    cy.contains('button', 'Update preferences').click()
    cy.get('[data-testid="appearance-language-fallback-sample"]').should(
      'contain',
      'Fallback sample text from English defaults.'
    )
  })
})
