describe('ui-foundation-accessibility', () => {
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
    cy.visit('/sign-in')
    cy.get('input[name="password"]').should('be.visible')
    assertVisibleIconButtonsHaveAccessibleName()
  })

  it('UI-FOUNDATION-ACCESSIBILITY-003 requires accessible naming for action controls in mobile header', () => {
    cy.viewport(375, 812)
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-accessibility@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    assertVisibleIconButtonsHaveAccessibleName()
  })
})
