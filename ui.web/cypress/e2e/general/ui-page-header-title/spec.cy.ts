describe('ui-page-header-title', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
  })

  function stubLocalSession(profileId: string) {
    expect(profileId).to.not.equal('')
    cy.intercept('POST', '**/api/auth/local/session', {
      statusCode: 200,
      body: {
        ok: true,
        session_token:
          'test-only-opaque-profile-bound-session-credential-000000000001',
      },
    }).as('pageHeaderLocalSession')
  }

  function signInTo(path: string) {
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.e2eSetSetupState('present')
      cy.e2eEnsureSignedOut()
      stubLocalSession(profile_id)
      cy.useBootstrappedProfile(profile_id, profile_name, { path })
    })
  }

  function assertCenteredHeader(testId: string, title: string) {
    cy.get(`[data-testid="${testId}-header-title"]`)
      .should('be.visible')
      .and('have.attr', 'data-centered', 'true')
      .and('contain', title)
    cy.get(`[data-testid="${testId}-page-icon"]`).should('be.visible')
    cy.get('header').should('not.contain', 'Active:')
    cy.get('header').should('not.contain', 'Collection:')
    cy.get('header').should('not.contain', 'Planning list')
  }

  function assertHeaderTitleDoesNotOverlapActions(testId: string) {
    cy.get(`[data-testid="${testId}-header-title"]`).then(($title) => {
      const titleRect = $title[0].getBoundingClientRect()

      cy.get(`[data-testid="${testId}-global-header-actions"]`).then(
        ($actions) => {
          const actionsRect = $actions[0].getBoundingClientRect()

          expect(
            titleRect.right,
            `${testId} title stays clear of header actions`
          ).to.be.lessThan(actionsRect.left - 8)
        }
      )
    })
  }

  function assertInventoryHeaderTitleBetweenSearchAndActions() {
    cy.get('[data-testid="inventory-header-title"]')
      .should('be.visible')
      .and('contain', 'Inventory')
    cy.get('[data-testid="inventory-page-icon"]').should('be.visible')
    cy.get('[data-testid="inventory-header-title"]').then(($title) => {
      const titleRect = $title[0].getBoundingClientRect()
      cy.contains('button', 'Search').then(($search) => {
        const searchRect = $search[0].getBoundingClientRect()
        expect(titleRect.left).to.be.greaterThan(searchRect.right)
      })
      cy.get('[data-testid="inventory-global-header-actions"]').then(
        ($actions) => {
          const actionsRect = $actions[0].getBoundingClientRect()
          expect(titleRect.right).to.be.lessThan(actionsRect.left - 8)
        }
      )
    })
  }

  it('UI-PAGE-HEADER-TITLE-001 keeps Inventory title visible between search and compact actions', () => {
    cy.viewport(1240, 720)

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    assertInventoryHeaderTitleBetweenSearchAndActions()
    cy.get('[data-testid="inventory-global-header-actions"]').should('be.visible')
    cy.get('header').should('not.contain', 'Active:')
    cy.get('header').should('not.contain', 'Collection:')
    cy.get('header').should('not.contain', 'Planning list')
  })

  it('UI-PAGE-HEADER-TITLE-002 centers primary page titles with icons and no inline context text', () => {
    cy.viewport(2048, 900)

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    assertInventoryHeaderTitleBetweenSearchAndActions()

    cy.visit('/collections/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/collections\/?$/)
    assertCenteredHeader('collections', 'Collections')
    assertHeaderTitleDoesNotOverlapActions('collections')

    cy.visit('/wishlist/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    assertCenteredHeader('wishlist', 'Wishlist')
    assertHeaderTitleDoesNotOverlapActions('wishlist')
  })
})
