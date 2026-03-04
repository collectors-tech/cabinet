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

  it('UI-FOUNDATION-ACCESSIBILITY-001 closes modal on Escape and restores focus to trigger', () => {
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
    cy.intercept('POST', '/api/ai/suggest/title', {
      statusCode: 200,
      body: {
        part_number: 'PN-A11Y-NEW',
        title: 'AI Suggested Item',
        confidence: 0.91,
      },
    }).as('aiSuggest')

    signInToInventory()
    cy.wait('@items')
    cy.wait('@profile')

    cy.get('[data-testid="inventory-ai-title-input"]').type('A11Y title input')
    cy.get('[data-testid="inventory-ai-suggest-title"]').click()
    cy.wait('@aiSuggest')

    cy.get('[data-testid="inventory-ai-apply"]').focus().click()
    cy.get('[data-testid="inventory-ai-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="inventory-ai-confirm-dialog"]').within(() => {
      cy.focused().should('exist')
    })
    cy.get('body').type('{esc}')
    cy.get('[data-testid="inventory-ai-confirm-dialog"]').should('not.exist')
    cy.get('[data-testid="inventory-ai-apply"]').should('be.focused')
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
    cy.get('input[placeholder="Filter by title or ID..."]')
      .focus()
      .type('missing-a11y-item{enter}')
    cy.contains('No results.').should('be.visible')
  })
})
