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

  it('UI-SCREEN-COLLECTIONS-026 opens and validates collection row side-panel workflows', () => {
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
    cy.request('/api/profiles/e2e-profile-001/settings').then((response) => {
      const settings = (response.body.settings ?? {}) as Record<string, string>
      const persisted = JSON.parse(
        settings['collections.workspace.v1'] ?? '{}'
      ) as {
        collections?: string[]
      }
      expect(persisted.collections).to.include('Store 2 Panel')
      expect(persisted.collections).not.to.include('Store 2')
    })
    cy.reload()
    cy.wait('@loadCollectionSettings')
    cy.get('[data-testid="collections-row-store-2-panel"]').scrollIntoView()
    cy.get('[data-testid="collections-row-name-store-2-panel"]').should(
      'contain.text',
      'Store 2 Panel'
    )
  })

  it('keeps the side panel open and skips persistence on duplicate rename', () => {
    signInToCollections()

    cy.get('@saveCollectionSettings.all').should('have.length', 0)
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )

    cy.get('[data-testid="collections-row-edit-store-1"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid="collections-edit-panel"]')
      .should('be.visible')
      .should('have.attr', 'data-side', 'right')
    cy.get('[data-testid="collections-edit-input"]')
      .should('have.value', 'Store 1')
      .clear()
      .type('Store 2')

    cy.get('[data-testid="collections-edit-submit"]').click()

    cy.get('[data-testid="collections-edit-panel"]').should('be.visible')
    cy.get('[data-testid="collections-edit-input"]').should(
      'have.value',
      'Store 2'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
    cy.get('[data-testid="collections-row-name-store-1"]').should(
      'contain.text',
      'Store 1'
    )
    cy.get('[data-testid="collections-row-name-store-2"]').should(
      'contain.text',
      'Store 2'
    )
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Store 1'
    )
  })
})
