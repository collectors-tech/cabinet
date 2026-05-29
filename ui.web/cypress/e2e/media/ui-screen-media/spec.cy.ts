describe('ui-screen-media', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eEnsureSignedOut()
  })

  it('UI-SCREEN-MEDIA-006 opens Media workspace shell from navigation', () => {
    cy.e2eBootstrap({ minimalProfile: true }).then((profile) =>
      cy.useBootstrappedProfile(profile.profile_id, profile.profile_name, {
        path: '/inventory/',
        shellWorkspace: 'navigation',
      })
    )
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    cy.get('[data-testid="sidebar-nav-link-media"]')
      .should('be.visible')
      .and('have.attr', 'href', '/media')
    cy.visit('/media/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/media\/?$/)
    cy.title().should('eq', 'Cabinet - Media')

    cy.get('[data-testid="media-workspace"]').should('be.visible')
    cy.get('[data-testid="media-header-title"]')
      .should('be.visible')
      .and('have.attr', 'data-centered', 'true')
      .and('contain', 'Media')
    cy.get('[data-testid="media-page-icon"]').should('be.visible')
    cy.get('[data-testid="media-card-grid"]')
      .should('be.visible')
      .find('[data-testid^="media-card-"]')
      .should('have.length', 3)

    cy.get('[data-testid="media-card-media-slot-car-front"]')
      .should('contain', 'AFX Mustang front view')
      .and('contain', 'Unlinked')
      .and('contain', 'Analysis ready')
      .and('contain', '92%')
    cy.get('[data-testid="media-open-media-slot-car-front"]').should(
      'be.enabled'
    )
    cy.get('[data-testid="media-analyze-media-slot-car-front"]').should(
      'be.disabled'
    )
    cy.get('[data-testid="media-assign-media-slot-car-front"]').should(
      'be.enabled'
    )
    cy.get('[data-testid="media-upload-action"]').should('be.disabled')
    cy.get('[data-testid="media-download-selected-action"]').should(
      'be.disabled'
    )

    cy.get('[data-testid="media-filter-unlinked"]').click()
    cy.get('[data-testid="media-card-grid"]')
      .find('[data-testid^="media-card-"]')
      .should('have.length', 1)
    cy.get('[data-testid="media-card-media-slot-car-front"]').should(
      'be.visible'
    )
    cy.get('[data-testid="media-card-media-porsche-box"]').should('not.exist')
  })
})
