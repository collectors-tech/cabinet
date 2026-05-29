describe('ui-screen-media', () => {
  const mediaResponse = {
    filter: 'all',
    summary: {
      total: 3,
      unlinked: 1,
      linked_inventory: 1,
      linked_wishlist: 1,
      linked_both: 0,
      ready_for_review: 2,
    },
    assets: [
      {
        id: 'media-slot-car-front',
        title: 'AFX Mustang front view',
        filename: 'slot-car-front.jpg',
        uploaded_at: '2026-05-25 10:20',
        linkage_state: 'unlinked',
        analysis_status: 'ready',
        source: 'Chat attachment',
        download_filename: 'slot-car-front-media-sl.jpg',
      },
      {
        id: 'media-porsche-box',
        title: 'Porsche 917 box side',
        filename: 'porsche-box.jpg',
        uploaded_at: '2026-05-24 16:45',
        linkage_state: 'linked_inventory',
        analysis_status: 'pending',
        source: 'Inventory photo',
        item_id: 'item-porsche',
        thumbnail_url: '/api/items/item-porsche/photos/media-porsche-box/file?variant=thumbnail',
        download_filename: 'porsche-917-box-side-media-po.jpg',
      },
      {
        id: 'media-wishlist-reference',
        title: 'Wanted chassis reference',
        filename: 'wishlist-reference.jpg',
        uploaded_at: '2026-05-23 08:12',
        linkage_state: 'linked_wishlist',
        analysis_status: 'not_analyzed',
        source: 'Wishlist evidence',
        wishlist_id: 'wish-chassis',
        download_filename: 'wanted-chassis-reference-media-wi.jpg',
      },
    ],
  }

  beforeEach(() => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eEnsureSignedOut()
  })

  function openMediaWorkspace() {
    cy.e2eBootstrap({ minimalProfile: true }).then((profile) =>
      cy.useBootstrappedProfile(profile.profile_id, profile.profile_name, {
        path: '/inventory/',
        shellWorkspace: 'navigation',
      })
    )
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('UI-SCREEN-MEDIA-006 opens Media workspace shell from navigation', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')
    cy.intercept('GET', '/api/media/assets?filter=unlinked', {
      statusCode: 200,
      body: {
        ...mediaResponse,
        filter: 'unlinked',
        assets: mediaResponse.assets.filter(
          (asset) => asset.linkage_state === 'unlinked'
        ),
      },
    }).as('unlinkedMediaAssets')

    openMediaWorkspace()

    cy.get('[data-testid="sidebar-nav-link-media"]')
      .should('be.visible')
      .and('have.attr', 'href', '/media')
    cy.visit('/media/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/media\/?$/)
    cy.title().should('eq', 'Cabinet - Media')
    cy.wait('@mediaAssets')

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
      .and('contain', 'slot-car-front-media-sl.jpg')
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
    cy.wait('@unlinkedMediaAssets')
    cy.get('[data-testid="media-card-grid"]')
      .find('[data-testid^="media-card-"]')
      .should('have.length', 1)
    cy.get('[data-testid="media-card-media-slot-car-front"]').should(
      'be.visible'
    )
    cy.get('[data-testid="media-card-media-porsche-box"]').should('not.exist')
  })

  it('UI-SCREEN-MEDIA-008 previews selected media download filenames from API state', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')
    cy.intercept('POST', '/api/media/downloads/preview', (req) => {
      expect(req.body).to.deep.equal({
        asset_ids: ['media-slot-car-front'],
        filter: 'all',
      })
      req.reply({
        statusCode: 200,
        body: {
          asset_ids: ['media-slot-car-front'],
          count: 1,
          filenames: ['slot-car-front-media-sl.jpg'],
          allowed: true,
        },
      })
    }).as('downloadPreview')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-select-media-slot-car-front"]').click()
    cy.get('[data-testid="media-download-selected-action"]')
      .should('be.enabled')
      .click()
    cy.wait('@downloadPreview')
    cy.get('[data-testid="media-download-preview"]')
      .should('be.visible')
      .and('contain', '1 file ready')
      .and('contain', 'slot-car-front-media-sl.jpg')
  })

  it('UI-SCREEN-MEDIA-011 confirms assignment and refreshes API-backed linkage state', () => {
    let assigned = false

    cy.intercept('GET', '/api/media/assets', (req) => {
      req.reply({
        statusCode: 200,
        body: assigned
          ? {
              ...mediaResponse,
              summary: {
                ...mediaResponse.summary,
                unlinked: 0,
                linked_wishlist: 2,
              },
              assets: mediaResponse.assets.map((asset) =>
                asset.id === 'media-slot-car-front'
                  ? {
                      ...asset,
                      linkage_state: 'linked_wishlist',
                      wishlist_id: 'wish-slot-car',
                    }
                  : asset
              ),
            }
          : mediaResponse,
      })
    }).as('mediaAssets')
    cy.intercept('POST', '/api/media/assignments/preview', (req) => {
      expect(req.body).to.deep.equal({
        asset_id: 'media-slot-car-front',
        target_type: 'wishlist',
        target_id: 'wish-slot-car',
      })
      req.reply({
        statusCode: 200,
        body: {
          asset_id: 'media-slot-car-front',
          target_type: 'wishlist',
          target_id: 'wish-slot-car',
          current_linkage_state: 'unlinked',
          projected_linkage_state: 'linked_wishlist',
          allowed: true,
          requires_confirmation: true,
          audit_summary:
            'Preserved media asset media-slot-car-front provenance while linking to wishlist target wish-slot-car.',
        },
      })
    }).as('assignmentPreview')
    cy.intercept('POST', '/api/media/assignments', (req) => {
      expect(req.body).to.deep.equal({
        asset_id: 'media-slot-car-front',
        target_type: 'wishlist',
        target_id: 'wish-slot-car',
      })
      assigned = true
      req.reply({
        statusCode: 200,
        body: {
          asset_id: 'media-slot-car-front',
          target_type: 'wishlist',
          target_id: 'wish-slot-car',
          current_linkage_state: 'linked_wishlist',
          projected_linkage_state: 'linked_wishlist',
          allowed: true,
          requires_confirmation: true,
          applied: true,
          audit_summary:
            'Preserved media asset media-slot-car-front provenance while linking to wishlist target wish-slot-car.',
        },
      })
    }).as('assignmentApply')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-assign-media-slot-car-front"]').click()
    cy.get('[data-testid="media-assignment-dialog"]').should('be.visible')
    cy.get('[data-testid="media-assignment-target-id"]').type('wish-slot-car')
    cy.get('[data-testid="media-assignment-preview-action"]').click()
    cy.wait('@assignmentPreview')
    cy.get('[data-testid="media-assignment-preview"]')
      .should('contain', 'Unlinked to Wishlist linked')
      .and('contain', 'Preserved media asset media-slot-car-front provenance')
    cy.get('[data-testid="media-assignment-confirm-action"]').click()
    cy.wait('@assignmentApply')
    cy.wait('@mediaAssets')
    cy.get('[data-testid="media-assignment-success"]')
      .should('be.visible')
      .and('contain', 'Preserved media asset media-slot-car-front provenance')
    cy.get('[data-testid="media-card-media-slot-car-front"]')
      .should('contain', 'Wishlist linked')
      .and('not.contain', 'Unlinked')
    cy.get('[data-testid="media-assign-media-slot-car-front"]').should(
      'be.disabled'
    )
  })

  it('UI-SCREEN-MEDIA-008 shows API empty, error, and retry states', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 500,
      body: { error: 'media_assets_unavailable' },
    }).as('mediaAssetsFailed')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssetsFailed')
    cy.get('[data-testid="media-error-state"]')
      .should('be.visible')
      .and('contain', 'media_assets_500')

    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: {
        filter: 'all',
        summary: {
          total: 0,
          unlinked: 0,
          linked_inventory: 0,
          linked_wishlist: 0,
          linked_both: 0,
          ready_for_review: 0,
        },
        assets: [],
      },
    }).as('mediaAssetsEmpty')
    cy.get('[data-testid="media-retry-action"]').click()
    cy.wait('@mediaAssetsEmpty')
    cy.get('[data-testid="media-empty-state"]')
      .should('be.visible')
      .and('contain', 'No media assets in this view')
  })
})
