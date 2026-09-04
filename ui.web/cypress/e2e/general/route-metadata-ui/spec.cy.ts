describe('route metadata UI', () => {
  function openRoute(path: string) {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path })
  }

  function assertHeaderTitleClearOfControls(testId: string) {
    cy.get(`[data-testid="${testId}"]`).then(($title) => {
      const titleRect = $title[0].getBoundingClientRect()
      expect(titleRect.left, `${testId} left edge`).to.be.gte(0)
      expect(titleRect.right, `${testId} right edge`).to.be.lte(
        Cypress.config('viewportWidth')
      )

      cy.get('[data-header-title-avoid="true"]').then(($controls) => {
        const controlsRect = $controls[0].getBoundingClientRect()
        const horizontallyOverlaps =
          titleRect.right > controlsRect.left - 8 &&
          titleRect.left < controlsRect.right + 8
        const verticallyOverlaps =
          titleRect.bottom > controlsRect.top &&
          titleRect.top < controlsRect.bottom

        expect(
          horizontallyOverlaps && verticallyOverlaps,
          `${testId} avoids header controls`
        ).to.eq(false)
      })
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-ROUTE-METADATA-006 renders canonical route headers and document titles without control overlap', () => {
    const routeCases = [
      {
        viewport: [1440, 900],
        path: '/settings/display',
        title: 'Display Settings',
        documentTitle: 'Cabinet - Display Settings',
        headerTestId: 'settings-header-title',
        iconTestId: 'settings-header-icon',
      },
      {
        viewport: [390, 844],
        path: '/scanner/',
        title: 'Market Watch',
        documentTitle: 'Cabinet - Market Watch',
        headerTestId: 'market-watch-header-title',
        iconTestId: 'market-watch-page-icon',
      },
      {
        viewport: [1280, 800],
        path: '/purchases/',
        title: 'Purchases',
        documentTitle: 'Cabinet - Purchases',
        headerTestId: 'purchases-header-title',
        iconTestId: 'purchases-page-icon',
      },
    ] as const

    cy.viewport(routeCases[0].viewport[0], routeCases[0].viewport[1])
    openRoute(routeCases[0].path)

    routeCases.forEach((routeCase, index) => {
      cy.viewport(routeCase.viewport[0], routeCase.viewport[1])
      if (index > 0) {
        cy.visit(routeCase.path)
      }
      cy.location('pathname', { timeout: 15000 }).should(
        'match',
        new RegExp(`^${routeCase.path.replace(/\/$/, '')}/?$`)
      )
      cy.title().should('eq', routeCase.documentTitle)
      cy.get(`[data-testid="${routeCase.headerTestId}"]`)
        .should('be.visible')
        .and('have.attr', 'data-centered', 'true')
        .and('contain', routeCase.title)
      cy.get(`[data-testid="${routeCase.iconTestId}"]`).should('be.visible')
      assertHeaderTitleClearOfControls(routeCase.headerTestId)
    })
  })
})
