describe('ui-screen-collections', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'

  const defaultCollectionInventoryItems = [
    {
      id: 'inventory-item-kobe-rookie',
      part_number: 'KOBE-1996',
      title: '1996 Topps Kobe Bryant rookie',
      status: 'active',
      category: 'Trading Card',
      brand: 'Topps',
      description: 'PSA candidate, Lakers lot',
    },
    {
      id: 'inventory-item-jordan-beam-team',
      part_number: 'JORDAN-1992',
      title: '1992 Beam Team Michael Jordan',
      status: 'active',
      category: 'Trading Card',
      brand: 'Stadium Club',
      description: 'Premium slab review queue',
    },
    {
      id: 'inventory-item-pikachu-shadowless',
      part_number: 'PKMN-SHADOWLESS',
      title: 'Shadowless Pikachu',
      status: 'active',
      category: 'Trading Card',
      brand: 'Pokemon',
      description: 'Binder copy ready for retail',
    },
    {
      id: 'inventory-item-charizard-base',
      part_number: 'PKMN-CHARIZARD',
      title: 'Base Set Charizard',
      status: 'active',
      category: 'Trading Card',
      brand: 'Pokemon',
      description: 'High-value vault candidate',
    },
    {
      id: 'inventory-item-metal-raiders-pack',
      part_number: 'YGO-METAL-RAIDERS',
      title: 'Yu-Gi-Oh Metal Raiders blister',
      status: 'active',
      category: 'Trading Card',
      brand: 'Konami',
      description: 'Sealed product awaiting assignment',
    },
  ]

  function mockDefaultCollectionInventoryItems() {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items: defaultCollectionInventoryItems },
    }).as('collectionsInventoryItems')
  }

  function mockCollectionInventoryItems(
    items: typeof defaultCollectionInventoryItems
  ) {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items },
    }).as('collectionsInventoryItems')
  }

  function signInToCollections(options: { mockInventory?: boolean } = {}) {
    const { mockInventory = true } = options
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    if (mockInventory) {
      mockDefaultCollectionInventoryItems()
    }
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/collections/' })
    })
    cy.wait('@loadCollectionSettings')
    if (mockInventory) {
      cy.wait('@collectionsInventoryItems')
    }
  }

  function bootstrapCollectionsProfile(path = '/collections/') {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    mockDefaultCollectionInventoryItems()
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path })
    })
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')
  }

  function collectionFilterOptionKey(value: string) {
    return value
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')
  }

  function selectWishlistCollection(collectionName: string) {
    cy.get('[data-testid="wishlist-table-collection-trigger"]').click()
    cy.get(
      `[data-testid="wishlist-table-collection-option-${collectionFilterOptionKey(collectionName)}"]`
    ).click()
  }

  function openWishlistCollectionFilter() {
    cy.get('[data-testid="wishlist-table-collection-trigger"]').click()
  }

  it('UI-SCREEN-COLLECTIONS-001 renders shared collections management table', () => {
    signInToCollections()

    cy.get('[data-testid="collections-section"]').should('be.visible')
    cy.get('[data-testid="collections-table-toolbar"]').should('be.visible')
    cy.get('[data-testid="collections-shared-table"]')
      .should('be.visible')
      .and('have.attr', 'data-table-surface', 'true')
      .find('table')
      .should('exist')
    cy.get('[data-testid="collections-table-pagination"]').should('be.visible')
    cy.contains('Collections table').should('not.exist')
    cy.contains(
      'Browse, create, rename, and remove collection rows from the same management surface.'
    ).should('not.exist')
    cy.get('[data-testid="collections-new-action"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-016 renders selected collection items in a lower table', () => {
    signInToCollections()

    cy.get('[data-testid="collections-workspace"]').should('be.visible')
    cy.get('[data-testid="collections-table-toolbar"]').should('be.visible')
    cy.get('[data-testid="collections-shared-table"]')
      .should('be.visible')
      .and('have.attr', 'data-table-surface', 'true')
      .find('table')
      .should('exist')
    cy.get('[data-testid="collections-members-table-toolbar"]').should('be.visible')
    cy.get('[data-testid="collections-members-table"]')
      .should('be.visible')
      .and('have.attr', 'data-table-surface', 'true')
      .find('table')
      .should('exist')
    cy.get('[data-testid="collections-members-table-pagination"]').should('be.visible')
    cy.get('[data-testid="collections-active-context"]').should('contain.text', 'All Items')
    cy.get('[data-testid="collections-selected-name"]').should('not.exist')
    cy.get('[data-testid="collections-assignment-panel"]').should('not.exist')
    cy.get('[data-testid="collections-member-row-inventory-item-kobe-rookie"]').should(
      'contain.text',
      '1996 Topps Kobe Bryant rookie'
    )
    cy.get('[data-testid="collections-member-row-inventory-item-charizard-base"]').should(
      'contain.text',
      'Base Set Charizard'
    )

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.get('[data-testid="collections-row-store-1"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('[data-testid="collections-active-context"]').should('contain.text', 'Store 1')
    cy.get('[data-testid="collections-member-row-inventory-item-pikachu-shadowless"]').should(
      'contain.text',
      'Shadowless Pikachu'
    )

    cy.get('[data-testid="collections-row-overflow"]').click()
    cy.get('[data-testid="collections-active-context"]').should('contain.text', 'Overflow')
    cy.get('[data-testid="collections-member-row-inventory-item-pikachu-shadowless"]').should(
      'not.exist'
    )
    cy.get('[data-testid="collections-members-empty-row"]').should(
      'contain.text',
      'No items are currently assigned to Overflow.'
    )
  })

  it('UI-SCREEN-COLLECTIONS-021 truncates long member table values instead of overflowing columns', () => {
    const longMemberName =
      'URL-BONZASLOTCARS-2026-MATCHBOX-HOT-WHEELS-TOMICA-AFX-AURORA-LONG-LONG-LONG'

    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    mockCollectionInventoryItems([
      {
        id: 'inventory-item-long-member-overflow',
        part_number:
          'PART-BONZASLOTCARS-002-MATCHBOX-HOT-WHEELS-MOOISYRQ-EXTRA-LONG',
        title: longMemberName,
        status: 'active',
        category: 'Slot Cars',
        brand: 'Bonza Slot Cars',
        description:
          'This deliberately long record verifies the members table clips text safely.',
      },
      ...defaultCollectionInventoryItems,
    ])
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/collections/' })
    })
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')

    cy.get('[data-testid="collections-members-table"]').then(($surface) => {
      const surface = $surface[0]
      expect(surface.scrollWidth, 'members table horizontal overflow').to.be.at.most(
        surface.clientWidth + 1
      )
    })
    cy.get('[data-testid="collections-member-row-inventory-item-long-member-overflow"]')
      .scrollIntoView()
      .should('be.visible')
      .within(() => {
        cy.get('td').each(($cell) => {
          expect($cell[0].scrollWidth, 'cell text overflow').to.be.at.most(
            $cell[0].clientWidth + 1
          )
        })
        cy.get('td')
          .first()
          .find('[data-testid^="collections-member-name-"]')
          .should('have.css', 'text-overflow', 'ellipsis')
      })
  })

  it('UI-SCREEN-COLLECTIONS-020 stretches both tables to the available viewport height', () => {
    signInToCollections()

    cy.get('[data-testid="collections-workspace"]').should('be.visible')
    cy.get('[data-testid="collections-section"]').should('be.visible')
    cy.get('[data-testid="collections-members-panel"]').should('be.visible')
    cy.get('[data-testid="collections-shared-table"]').should('be.visible')
    cy.get('[data-testid="collections-members-table"]').should('be.visible')

    cy.get('[data-testid="collections-section"]').then(($collectionsCard) => {
      cy.get('[data-testid="collections-members-panel"]').then(($membersCard) => {
        expect(
          $collectionsCard[0].getBoundingClientRect().height,
          'collections table card height'
        ).to.be.greaterThan(300)
        expect(
          $membersCard[0].getBoundingClientRect().height,
          'collection members card height'
        ).to.be.greaterThan(300)
      })
    })
    cy.get('[data-testid="collections-shared-table"]').then(($surface) => {
      expect(
        $surface[0].getBoundingClientRect().height,
        'collections table surface height'
      ).to.be.greaterThan(180)
    })
    cy.get('[data-testid="collections-members-table"]').then(($surface) => {
      expect(
        $surface[0].getBoundingClientRect().height,
        'members table surface height'
      ).to.be.greaterThan(180)
    })
    cy.window().then((win) => {
      expect(win.document.documentElement.scrollHeight).to.be.at.most(
        win.innerHeight + 2
      )
    })
  })

  it('UI-SCREEN-COLLECTIONS-018 keeps All Items count aligned with inventory members', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: Array.from({ length: 12 }, (_, index) => ({
          id: `collection-live-item-${index + 1}`,
          part_number: `COL-LIVE-${String(index + 1).padStart(3, '0')}`,
          title: `Collection Live Item ${index + 1}`,
          status: 'active',
          category: index % 2 === 0 ? 'Slot Car' : 'Trading Card',
          brand: index % 2 === 0 ? 'AFX' : 'Topps',
          description: `Live inventory item ${index + 1}`,
        })),
      },
    }).as('collectionsInventoryItems')

    signInToCollections({ mockInventory: false })
    cy.wait('@collectionsInventoryItems')

    cy.get('[data-testid="collections-row-count-all-items"]').should(
      'have.text',
      '12'
    )
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 12 of 12 items.'
    )
    cy.get('[data-testid="collections-member-row-collection-live-item-1"]')
      .scrollIntoView()
      .should('be.visible')
      .and('contain.text', 'Collection Live Item 1')
    cy.get('[data-testid="collections-member-row-collection-live-item-10"]')
      .scrollIntoView()
      .should('be.visible')
      .and('contain.text', 'Collection Live Item 10')
  })

  it('UI-SCREEN-COLLECTIONS-002 selects a row and persists active context across refresh', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSelection')

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.wait('@saveCollectionSelection')
    cy.get('[data-testid="collections-row-watch-list"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('[data-testid="collections-active-context"]').should('contain.text', 'Watch List')
    cy.get('[data-testid="collections-selected-name"]').should('not.exist')

    cy.reload()
    cy.get('[data-testid="collections-active-context"]').should('contain.text', 'Watch List')
    cy.get('[data-testid="collections-row-watch-list"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
  })

  it('UI-SCREEN-COLLECTIONS-010 moves Browse into row-level View actions', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSelection')

    cy.get('[data-testid="collections-active-browse-control"]').should('not.exist')
    cy.get('[data-testid="collections-active-browse"]').should('not.exist')
    cy.get('[data-testid="collections-row-watch-list"]').scrollIntoView()
    cy.get('[data-testid="collections-row-edit-watch-list"]').should('be.visible')
    cy.get('[data-testid="collections-row-delete-watch-list"]').should('be.visible')
    cy.get('[data-testid="collections-row-view-watch-list"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'View Watch List in inventory')
      .click()
    cy.wait('@saveCollectionSelection')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get('[data-testid="collection-active-context"]').should(
      'contain.text',
      'Watch List'
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
    cy.get('[data-testid="collections-active-context"]').should('contain.text', 'Collections Alpha')
  })

  it('UI-SCREEN-COLLECTIONS-004 renames a collection from the row workflow and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-2"]').click()
    cy.get('[data-testid="collections-row-edit-store-2"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-edit-input"]').clear().type('Store 2 Prime')
    cy.get('[data-testid="collections-edit-submit"]').click()

    cy.contains('Store 2 renamed to Store 2 Prime.').should('be.visible')
    cy.get('[data-testid="collections-row-store-2-prime"]')
      .scrollIntoView()
      .should('exist')
    cy.reload()
    cy.get('[data-testid="collections-row-store-2-prime"]')
      .scrollIntoView()
      .should('exist')
  })

  it('UI-SCREEN-COLLECTIONS-005 deletes a collection and releases assigned items', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.get('[data-testid="collections-member-shadowless-pikachu"]').should('be.visible')
    cy.get('[data-testid="collections-row-delete-store-1"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-delete-submit"]').click()

    cy.contains('Store 1 removed from workspace collections.').should('be.visible')
    cy.get('[data-testid="collections-row-store-1"]').should('not.exist')
    cy.get('[data-testid="collections-active-context"]').should('contain.text', 'All Items')
    cy.get('[data-testid="collections-member-row-inventory-item-pikachu-shadowless"]').should(
      'contain.text',
      'Currently in Unassigned.'
    )
  })

  it('UI-SCREEN-COLLECTIONS-006 filters collections within the shared table surface', () => {
    signInToCollections()

    cy.get('[data-testid="collections-management-summary"]').should('contain.text', 'Showing 6 of 6 collections.')
    cy.get('[data-testid="collections-search-input"]').type('watch')
    cy.get('[data-testid="collections-row-watch-list"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('not.exist')
    cy.get('[data-testid="collections-management-summary"]').should('contain.text', 'Showing 1 of 6 collections.')
  })

  it('UI-SCREEN-COLLECTIONS-017 filters collection members within the members table surface', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-all-items"]').click()
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 5 of 5 items.'
    )
    cy.get('[data-testid="collections-members-search-input"]').type('charizard')
    cy.get('[data-testid="collections-member-row-inventory-item-charizard-base"]').should(
      'exist'
    )
    cy.get('[data-testid="collections-member-row-inventory-item-kobe-rookie"]').should(
      'not.exist'
    )
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 1 of 5 items.'
    )
  })

  it('UI-SCREEN-COLLECTIONS-007 keeps item assignment actions out of Collections', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
    cy.get('[data-testid="collections-assignment-panel"]').should('not.exist')
    cy.get('[data-testid^="collections-move-target-"]').should('not.exist')
    cy.get('[data-testid^="collections-move-submit-"]').should('not.exist')
    cy.get('[data-testid^="collections-unassign-submit-"]').should('not.exist')
  })

  it('UI-SCREEN-COLLECTIONS-009 retains tag iconography for collections route identity', () => {
    signInToCollections()

    cy.get('[data-testid="sidebar-nav-link-collections"]').should('be.visible')
    cy.get('[data-testid="collections-page-icon"]').should('exist')
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

    cy.request('/api/profiles/e2e-profile-001/settings').then((response) => {
      const settings = (response.body.settings ?? {}) as Record<string, string>
      const persisted = JSON.parse(settings[collectionsSettingsKey] ?? '{}') as {
        collections?: string[]
        activeCollection?: string
      }

      expect(persisted.collections).to.include('Profile Persisted Vault')
      expect(persisted.activeCollection).to.equal('Profile Persisted Vault')
    })

    cy.reload()
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Profile Persisted Vault'
    )
    cy.get('[data-testid="collections-members-empty-row"]').should(
      'contain.text',
      'No items are currently assigned to Profile Persisted Vault.'
    )
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

          cy.get('[data-testid="collections-row-profile-two-vault"]')
            .scrollIntoView()
            .should('exist')
          cy.get('[data-testid="collections-active-context"]').should(
            'contain.text',
            'Profile Two Vault'
          )
          cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should(
            'exist'
          )
        }
      )
    })
  })

  it('UI-SCREEN-COLLECTIONS-012 keeps inventory collection create in the folder tree', () => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/inventory/' })
    })
    cy.wait('@loadCollectionSettings')

    cy.get('[data-testid="folder-tree-add-root"]').click()
    cy.get('[data-testid="folder-tree-name-input"]').type('Inventory Sync Shelf')
    cy.get('[data-testid="folder-tree-create-submit"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="collection-active-context"]').should(
      'contain.text',
      'Inventory Sync Shelf'
    )

    cy.reload()
    cy.wait('@loadCollectionSettings')
    cy.get('[data-testid="collection-active-context"]').should(
      'contain.text',
      'Inventory Sync Shelf'
    )
  })

  it('UI-SCREEN-COLLECTIONS-013 reflects wishlist table collection create inside the collections manager', () => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/*/settings').as('loadCollectionSettings')
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/wishlist/' })
    })
    cy.wait('@loadCollectionSettings')

    cy.get('[data-testid="wishlist-create-collection-action"]').click()
    cy.get('[data-testid="wishlist-create-collection-name"]').type('Wishlist Sync Shelf')
    cy.get('[data-testid="wishlist-create-collection-save"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="wishlist-table-add-collection"]').should('not.exist')
    cy.get('[data-testid="wishlist-table-collection-selected"]').should('not.exist')

    cy.visit('/collections/')
    cy.wait('@loadCollectionSettings')
    cy.get('[data-testid="collections-row-wishlist-sync-shelf"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Wishlist Sync Shelf'
    )
  })

  it('UI-SCREEN-COLLECTIONS-014 propagates rename into the wishlist table collection filter', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-2"]').click()
    cy.get('[data-testid="collections-row-edit-store-2"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-edit-input"]').clear().type('Store 2 Routed')
    cy.get('[data-testid="collections-edit-submit"]').click()
    cy.contains('Store 2 renamed to Store 2 Routed.').should('be.visible')

    cy.visit('/wishlist/')
    selectWishlistCollection('Store 2 Routed')
    cy.get('[data-testid="wishlist-table-collection-trigger"]').should(
      'contain.text',
      'Store 2 Routed'
    )
    openWishlistCollectionFilter()
    cy.get('[data-testid="wishlist-table-collection-option-store-2-routed"]').should(
      'exist'
    )
    cy.get('[data-testid="wishlist-table-collection-option-store-2"]').should(
      'not.exist'
    )
  })

  it('UI-SCREEN-COLLECTIONS-015 removes deleted collections from compact filters', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as('saveCollectionSettings')

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.wait('@saveCollectionSettings').then(({ request }) => {
      const settings = (request.body.settings ?? {}) as Record<string, string>
      const persisted = JSON.parse(settings[collectionsSettingsKey] ?? '{}') as {
        activeCollection?: string
      }
      expect(persisted.activeCollection).to.equal('Store 1')
    })
    cy.get('[data-testid="collections-row-delete-store-1"]').scrollIntoView().click({ force: true })
    cy.get('[data-testid="collections-delete-submit"]').click()
    cy.contains('Store 1 removed from workspace collections.').should('be.visible')
    cy.wait('@saveCollectionSettings').then(({ request }) => {
      const settings = (request.body.settings ?? {}) as Record<string, string>
      const persisted = JSON.parse(settings[collectionsSettingsKey] ?? '{}') as {
        collections?: string[]
        activeCollection?: string
      }
      expect(persisted.activeCollection).to.equal('All Items')
      expect(
        persisted.collections ?? [],
        `delete request collections: ${JSON.stringify(persisted.collections ?? [])}`
      ).not.to.include('Store 1')
    })
    cy.request('/api/profiles/e2e-profile-001/settings').then((response) => {
      const settings = (response.body.settings ?? {}) as Record<string, string>
      const persisted = JSON.parse(settings[collectionsSettingsKey] ?? '{}') as {
        collections?: string[]
        activeCollection?: string
      }
      expect(persisted.activeCollection).to.equal('All Items')
      expect(
        persisted.collections ?? [],
        `persisted collections: ${JSON.stringify(persisted.collections ?? [])}`
      ).not.to.include('Store 1')
    })

    cy.visit('/wishlist/')
    cy.wait('@loadCollectionSettings')
    cy.get('[data-testid="wishlist-table-collection-selected"]').should('not.exist')
    openWishlistCollectionFilter()
    cy.get('[data-testid="wishlist-table-collection-option-store-1"]').should(
      'not.exist'
    )
  })
})
