describe('language-switch-ui', () => {
  function enterHomeWithLocalSession() {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.e2eEnsureSignedOut()
      cy.stubLocalServerSession(profile_id)
      cy.setCookie('sidebar_state', 'true')
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/dashboard',
      })
    })
  }

  function switchLanguage(code: 'en' | 'zh' | 'ja') {
    cy.get('[data-testid="header-language-switch-trigger"]')
      .filter(':visible')
      .click()
    cy.get(`[data-testid="header-language-option-${code}"]`)
      .filter(':visible')
      .click()
  }

  it('UI-FOUNDATION-THEME-RTL-I18N-004 updates visible shell and page labels when language is changed', () => {
    enterHomeWithLocalSession()

    cy.get('[data-testid="header-language-switch-trigger"]').should('contain', 'EN')
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should('contain', 'Home')
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
