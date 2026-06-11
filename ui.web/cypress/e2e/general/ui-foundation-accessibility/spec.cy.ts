describe('ui-foundation-accessibility', () => {
  function signInToInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  function assertVisibleIconButtonsHaveAccessibleName() {
    cy.get('button:visible').each(($button) => {
      const hasSvgIcon = $button.find('svg').length > 0
      if (!hasSvgIcon) {
        return
      }
      const visibleText = ($button.text() ?? '').replace(/\s+/g, ' ').trim()
      const ariaLabel = ($button.attr('aria-label') ?? '').trim()
      const title = ($button.attr('title') ?? '').trim()
      const srOnlyText = ($button.find('.sr-only').text() ?? '').replace(/\s+/g, ' ').trim()
      const accessibleName = ariaLabel || title || visibleText || srOnlyText
      expect(
        accessibleName,
        `icon action button must have accessible name: ${$button.prop('outerHTML')}`
      ).to.not.equal('')
    })
  }

  it('UI-FOUNDATION-ACCESSIBILITY-003 requires accessible naming for action controls on sign-in screen', () => {
    cy.e2eSetSetupState('present')
    cy.visit('/sign-in')
    assertVisibleIconButtonsHaveAccessibleName()
  })

  it('UI-FOUNDATION-ACCESSIBILITY-003 requires accessible naming for action controls in mobile header', () => {
    cy.viewport(375, 812)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    assertVisibleIconButtonsHaveAccessibleName()
  })

  it('UI-FOUNDATION-ACCESSIBILITY-001 closes modal on Escape without trapping focus', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-a11y-1',
            part_number: 'PN-A11Y-1',
            title: 'Accessibility Item',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-a11y' },
    }).as('profile')
    signInToInventory()
    cy.wait('@items')
    cy.wait('@profile')

    cy.get('[data-testid="inventory-new-action"]').focus().click()
    cy.get('[data-testid="inventory-item-create-dialog"]').should('be.visible')
    cy.get('[data-testid="inventory-item-create-dialog"]').within(() => {
      cy.focused().should('exist')
    })
    cy.get('body').type('{esc}')
    cy.get('[data-testid="inventory-item-create-dialog"]').should('not.exist')
    cy.get('[data-testid="inventory-new-action"]').should('be.visible')
  })

  it('UI-FOUNDATION-ACCESSIBILITY-002 supports keyboard-only execution for inventory workflow controls', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-a11y-kbd',
            part_number: 'PN-A11Y-KBD',
            title: 'Keyboard Flow Item',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('itemsKeyboard')

    signInToInventory()
    cy.wait('@itemsKeyboard')

    cy.get('main').should('be.visible')
    cy.get('button[aria-label="Switch to cards view"]').focus().type('{enter}')
    cy.contains('Status:').should('be.visible')
    cy.get('button[aria-label="Switch to rows view"]').focus().type('{enter}')
    cy.get('table').should('be.visible')
    cy.get('[data-testid="inventory-table-search-input"]')
      .should('be.visible')
      .focus()
      .type('missing-a11y-item{enter}')
    cy.contains('No results.').should('be.visible')
  })
})
