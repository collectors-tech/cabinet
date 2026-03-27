describe('language-switch-ui', () => {
  function signInToHome() {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.visit('/sign-in?redirect=%2F')
    cy.get('input[name="email"]').clear().type('e2e-theme-i18n@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('eq', '/')
  }

  function switchLanguage(code: 'en' | 'zh' | 'ja') {
    cy.get('[data-testid="header-language-switch-trigger"]').click()
    cy.get(`[data-testid="header-language-option-${code}"]`).click()
  }

  it('UI-FOUNDATION-THEME-RTL-I18N-004 updates visible shell and page labels when language is changed', () => {
    signInToHome()

    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should('contain', 'Dashboard')
    cy.contains('h1', 'Home').should('be.visible')

    switchLanguage('zh')

    cy.get('[data-testid="header-language-switch-trigger"]').should('contain', 'ZH')
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should('contain', '首页')
    cy.get('[data-testid="sidebar-nav-link-inventory"]').should('contain', '库存')
    cy.contains('h1', '首页').should('be.visible')

    cy.get('[data-testid="sidebar-nav-link-inventory"]').click()
    cy.location('pathname').should('eq', '/inventory')
    cy.contains('h1', '库存').should('be.visible')

    switchLanguage('ja')

    cy.get('[data-testid="header-language-switch-trigger"]').should('contain', 'JA')
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should('contain', 'ホーム')
    cy.get('[data-testid="sidebar-nav-link-inventory"]').should('contain', '在庫')
    cy.contains('h1', '在庫').should('be.visible')
  })
})
