describe('ui-screen-collections', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'

  function signInToCollections() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/collections/' })
    })
    cy.wait('@loadCollectionSettings')
  }

  function bootstrapCollectionsProfile(path = '/collections/') {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path })
    })
    cy.wait('@loadCollectionSettings')
  }

  it('UI-SCREEN-COLLECTIONS-001 renders shared collections management table', () => {
    signInToCollections()

    cy.get('[data-testid="collections-section"]').should('be.visible')
    cy.get('[data-testid="collections-shared-table"]').should('be.visible')
    cy.get('[data-testid="collections-new-action"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-002 selects a row and persists active context across refresh', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSelection')

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.wait('@saveCollectionSelection')
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Watch List')
    cy.get('[data-testid="collections-active-context-message"]').should(
      'contain.text',
      'Active collection is Watch List'
    )

    cy.reload()
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Watch List')
    cy.get('[data-testid="collections-active-context-persistence"]').should(
      'contain.text',
      'Persists for this signed-in profile'
    )
  })

  it('UI-SCREEN-COLLECTIONS-003 creates a collection from the table workflow and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-new-action"]').click()
    cy.get('[data-testid="collections-create-input"]').type('Collections Alpha')
    cy.get('[data-testid="collections-create-submit"]').click()

    cy.contains('Collections Alpha created and set as the active collection.').should('be.visible')
    cy.get('[data-testid="collections-row-collections-alpha"]').should('be.visible')
    cy.reload()
    cy.get('[data-testid="collections-row-collections-alpha"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Collections Alpha')
  })

  it('UI-SCREEN-COLLECTIONS-004 renames a collection from the row workflow and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-2"]').click()
    cy.get('[data-testid="collections-row-edit-store-2"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-edit-input"]').clear().type('Store 2 Prime')
    cy.get('[data-testid="collections-edit-submit"]').click()

    cy.contains('Store 2 renamed to Store 2 Prime.').should('be.visible')
    cy.get('[data-testid="collections-row-store-2-prime"]').should('be.visible')
    cy.reload()
    cy.get('[data-testid="collections-row-store-2-prime"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-005 deletes a collection and releases assigned items', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.get('[data-testid="collections-member-shadowless-pikachu"]').should('be.visible')
    cy.get('[data-testid="collections-row-delete-store-1"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-delete-submit"]').click()

    cy.contains('Store 1 removed from workspace collections.').should('be.visible')
    cy.get('[data-testid="collections-row-store-1"]').should('not.exist')

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.get('[data-testid="collections-assignment-select"]').click()
    cy.contains('Shadowless Pikachu (Unassigned)').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-006 filters collections within the shared table surface', () => {
    signInToCollections()

    cy.get('[data-testid="collections-management-summary"]').should('contain.text', 'Showing 6 of 6 collections.')
    cy.get('[data-testid="collections-search-input"]').type('watch')
    cy.get('[data-testid="collections-row-watch-list"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('not.exist')
    cy.get('[data-testid="collections-management-summary"]').should('contain.text', 'Showing 1 of 6 collections.')
  })

  it('UI-SCREEN-COLLECTIONS-007 assigns an item into the selected collection and persists after refresh', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')

    cy.get('[data-testid="collections-row-warehouse-1"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="collections-assignment-select"]').click()
    cy.contains('Base Set Charizard (Unassigned)').click()
    cy.get('[data-testid="collections-assignment-submit"]').click()
    cy.wait('@saveCollectionSettings')

    cy.contains('Base Set Charizard assigned to Warehouse 1.').should('be.visible')
    cy.get('[data-testid="collections-member-base-set-charizard"]').should('be.visible')

    cy.reload()
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Warehouse 1')
    cy.get('[data-testid="collections-member-base-set-charizard"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-008 moves an assigned item between collections and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
    cy.get('[data-testid="collections-move-target-1996-topps-kobe-bryant-rookie"]').click()
    cy.contains('[role="option"]', 'Warehouse 1').click()
    cy.get('[data-testid="collections-move-submit-1996-topps-kobe-bryant-rookie"]').click()

    cy.contains('1996 Topps Kobe Bryant rookie moved to Warehouse 1.').should('be.visible')
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('not.exist')

    cy.get('[data-testid="collections-row-warehouse-1"]').click()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
    cy.reload()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-008A unassigns an item directly from the selected collection and persists after refresh', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
    cy.get('[data-testid="collections-unassign-submit-1996-topps-kobe-bryant-rookie"]').click()
    cy.wait('@saveCollectionSettings')

    cy.contains('1996 Topps Kobe Bryant rookie removed from Watch List.').should('be.visible')
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('not.exist')

    cy.get('[data-testid="collections-assignment-select"]').click()
    cy.contains('1996 Topps Kobe Bryant rookie (Unassigned)').should('be.visible')

    cy.reload()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('not.exist')
    cy.get('[data-testid="collections-assignment-select"]').click()
    cy.contains('1996 Topps Kobe Bryant rookie (Unassigned)').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-009 retains tag iconography for collections route identity', () => {
    signInToCollections()

    cy.get('[data-testid="sidebar-nav-link-collections"]').should('be.visible')
    cy.get('[data-testid="collections-page-icon"]').should('be.visible')
    cy.get('[data-testid="sidebar-nav-link-collections"] svg').should('have.class', 'lucide-tag')
    cy.get('[data-testid="collections-page-icon"]').should('have.class', 'lucide-tag')
  })

  it('UI-SCREEN-COLLECTIONS-010 persists collection state through profile settings and across refresh', () => {
    bootstrapCollectionsProfile()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')

    cy.get('[data-testid="collections-new-action"]').click()
    cy.get('[data-testid="collections-create-input"]').type('Profile Persisted Vault')
    cy.get('[data-testid="collections-create-submit"]').click()
    cy.wait('@saveCollectionSettings')

    cy.get('[data-testid="collections-row-profile-persisted-vault"]').click()
    cy.get('[data-testid="collections-assignment-select"]').click()
    cy.contains('Base Set Charizard (Unassigned)').click()
    cy.get('[data-testid="collections-assignment-submit"]').click()
    cy.wait('@saveCollectionSettings')

    cy.request('/api/profiles/e2e-profile-001/settings').then((response) => {
      const settings = (response.body.settings ?? {}) as Record<string, string>
      const persisted = JSON.parse(settings[collectionsSettingsKey] ?? '{}') as {
        collections?: string[]
        activeCollection?: string
        items?: Array<{ id?: string; collectionName?: string | null }>
      }

      expect(persisted.collections).to.include('Profile Persisted Vault')
      expect(persisted.activeCollection).to.equal('Profile Persisted Vault')
      expect(persisted.items).to.satisfy(
        (items?: Array<{ id?: string; collectionName?: string | null }>) =>
          Array.isArray(items) &&
          items.some(
            (item) =>
              item.id === 'inventory-item-charizard-base' &&
              item.collectionName === 'Profile Persisted Vault'
          )
      )
    })

    cy.reload()
    cy.get('[data-testid="collections-selected-name"]').should(
      'contain.text',
      'Profile Persisted Vault'
    )
    cy.get('[data-testid="collections-member-base-set-charizard"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-011 switches collection state with the active profile', () => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.request('POST', '/api/profiles', { name: 'Collections Profile Two' }).then(
        (createResponse) => {
          expect(createResponse.status).to.be.oneOf([200, 201])
          const secondProfileId = createResponse.body.id as string

          const secondProfileSettings = {
            [collectionsSettingsKey]: JSON.stringify({
              collections: ['All Items', 'Profile Two Vault'],
              activeCollection: 'Profile Two Vault',
              items: [
                {
                  id: 'inventory-item-kobe-rookie',
                  name: '1996 Topps Kobe Bryant rookie',
                  detail: 'PSA candidate, Lakers lot',
                  collectionName: 'Profile Two Vault',
                },
              ],
            }),
          }

          cy.request('PUT', `/api/profiles/${secondProfileId}/settings`, {
            settings: secondProfileSettings,
          }).its('status').should('eq', 200)

          cy.useBootstrappedProfile(profile_id, profile_name, { path: '/collections/' })
          cy.get('[data-testid="collections-row-profile-two-vault"]').should('not.exist')

          cy.request('PUT', '/api/profiles/active', { profile_id: secondProfileId })
            .its('status')
            .should('eq', 200)
          cy.visit('/collections/')

          cy.get('[data-testid="collections-row-profile-two-vault"]').should('be.visible')
          cy.get('[data-testid="collections-selected-name"]').should(
            'contain.text',
            'Profile Two Vault'
          )
          cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should(
            'be.visible'
          )
        }
      )
    })
  })

  it('UI-SCREEN-COLLECTIONS-012 reflects inventory inline collection create inside the collections manager', () => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/inventory/' })
    })
    cy.wait('@loadCollectionSettings')

    cy.get('[data-testid="collection-inline-add-new"]').click()
    cy.get('[data-testid="collection-inline-new-name"]').type('Inventory Sync Shelf')
    cy.get('[data-testid="collection-inline-save"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="collection-inline-picker-selected"]').should(
      'contain.text',
      'Inventory Sync Shelf'
    )

    cy.visit('/collections/')
    cy.wait('@loadCollectionSettings')
    cy.get('[data-testid="collections-row-inventory-sync-shelf"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should(
      'contain.text',
      'Inventory Sync Shelf'
    )
  })

  it('UI-SCREEN-COLLECTIONS-013 reflects wishlist inline collection create inside the collections manager', () => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/wishlist/' })
    })
    cy.wait('@loadCollectionSettings')

    cy.get('[data-testid="wishlist-inline-add-new"]').click()
    cy.get('[data-testid="wishlist-inline-new-name"]').type('Wishlist Sync Shelf')
    cy.get('[data-testid="wishlist-inline-save"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="wishlist-inline-picker-selected"]').should(
      'contain.text',
      'Wishlist Sync Shelf'
    )

    cy.visit('/collections/')
    cy.wait('@loadCollectionSettings')
    cy.get('[data-testid="collections-row-wishlist-sync-shelf"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should(
      'contain.text',
      'Wishlist Sync Shelf'
    )
  })

  it('UI-SCREEN-COLLECTIONS-014 propagates rename into inventory and wishlist pickers', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-2"]').click()
    cy.get('[data-testid="collections-row-edit-store-2"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-edit-input"]').clear().type('Store 2 Routed')
    cy.get('[data-testid="collections-edit-submit"]').click()
    cy.contains('Store 2 renamed to Store 2 Routed.').should('be.visible')

    cy.visit('/inventory/')
    cy.get('[data-testid="collection-inline-picker-option-store-2-routed"]').should('be.visible')
    cy.get('[data-testid="collection-inline-picker-option-store-2"]').should('not.exist')

    cy.visit('/wishlist/')
    cy.get('[data-testid="wishlist-inline-picker-option-store-2-routed"]').should('be.visible')
    cy.get('[data-testid="wishlist-inline-picker-option-store-2"]').should('not.exist')
  })

  it('UI-SCREEN-COLLECTIONS-015 propagates delete fallback into inventory and wishlist pickers', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="collections-row-delete-store-1"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-delete-submit"]').click()
    cy.contains('Store 1 removed from workspace collections.').should('be.visible')

    cy.visit('/inventory/')
    cy.get('[data-testid="collection-inline-picker-selected"]').should('contain.text', 'All Items')
    cy.get('[data-testid="collection-inline-picker-option-store-1"]').should('not.exist')

    cy.visit('/wishlist/')
    cy.get('[data-testid="wishlist-inline-picker-selected"]').should('contain.text', 'All Items')
    cy.get('[data-testid="wishlist-inline-picker-option-store-1"]').should('not.exist')
  })
})
