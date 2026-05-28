describe('collections-row-side-panel', () => {
  function signInToCollections() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/*/settings').as(
      'loadCollectionSettings'
    )
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as(
      'saveCollectionSettings'
    )
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/collections/',
      })
    })
    cy.wait('@loadCollectionSettings')
  }

  it('opens a right-side edit panel on double click and navigates visible records', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.get('[data-testid="collections-edit-panel"]').should('not.exist')
    cy.get('[data-testid="collections-edit-dialog"]').should('not.exist')

    cy.get('[data-testid="collections-row-store-1"]').dblclick()
    cy.get('[data-testid="collections-edit-panel"]')
      .should('be.visible')
      .should('have.attr', 'data-side', 'right')
    cy.get('[data-testid="collections-edit-input"]').should(
      'have.value',
      'Store 1'
    )

    cy.get('[data-testid="collections-edit-next"]').click()
    cy.get('[data-testid="collections-edit-input"]').should(
      'have.value',
      'Store 2'
    )
    cy.get('[data-testid="collections-edit-previous"]').click()
    cy.get('[data-testid="collections-edit-input"]').should(
      'have.value',
      'Store 1'
    )
    cy.get('[data-testid="collections-edit-next"]').click()

    cy.get('[data-testid="collections-edit-input"]')
      .clear()
      .type('Store 2 Panel')
    cy.get('[data-testid="collections-edit-submit"]').click()
    cy.wait('@saveCollectionSettings')

    cy.contains('Store 2 renamed to Store 2 Panel.').should('be.visible')
    cy.get('[data-testid="collections-row-store-2-panel"]').scrollIntoView()
    cy.get('[data-testid="collections-row-store-2-panel"]').should('exist')
    cy.get('[data-testid="collections-row-name-store-2-panel"]').should(
      'contain.text',
      'Store 2 Panel'
    )
    cy.reload()
    cy.wait('@loadCollectionSettings')
    cy.get('[data-testid="collections-row-store-2-panel"]').scrollIntoView()
    cy.get('[data-testid="collections-row-name-store-2-panel"]').should(
      'contain.text',
      'Store 2 Panel'
    )
  })
})
