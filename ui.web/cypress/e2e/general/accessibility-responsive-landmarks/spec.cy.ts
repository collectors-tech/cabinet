describe('accessibility-responsive-landmarks', () => {
  function signInToInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  function assertResponsiveLandmarks(pageTitle: string, viewportWidth: number) {
    cy.get('main:visible').should('have.length', 1)
    cy.get('[data-testid="inventory-header-title"]')
      .and('contain', pageTitle)
      .and('have.attr', 'aria-label')
      .and('include', pageTitle)
    if (viewportWidth >= 768) {
      cy.get('[data-testid="inventory-header-title"]').should('be.visible')
    }
    cy.get('header:visible').should('have.length.at.least', 1)
    cy.window().then((win) => {
      const doc = win.document.documentElement
      expect(doc.scrollWidth, 'document avoids horizontal overflow').to.be.lte(
        win.innerWidth + 1
      )
    })
  }

  it('UI-FOUNDATION-ACCESSIBILITY-004 preserves one main landmark and page heading across responsive shell widths', () => {
    const responsiveViewports = [
      { width: 1440, height: 900 },
      { width: 390, height: 844 },
    ]

    responsiveViewports.forEach((viewport) => {
      cy.viewport(viewport.width, viewport.height)
      signInToInventory()
      assertResponsiveLandmarks('Inventory', viewport.width)
    })
  })
})
