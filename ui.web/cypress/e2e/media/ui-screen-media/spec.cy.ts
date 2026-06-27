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
        thumbnail_url: '/api/media/assets/media-slot-car-front/file?variant=thumbnail',
        thumbnail_variations: ['Original', 'Thumbnail', 'Review crop'],
        notes: 'Initial intake note',
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

  function chooseMediaLinkageFilter(value: 'all' | 'unlinked') {
    cy.get('[data-testid="media-linkage-filter-trigger"]').click()
    cy.get(`[data-testid="media-filter-${value}"]`).click()
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
    cy.get('[data-testid="media-workspace"]').within(() => {
      cy.contains('h1', /^Media$/).should('not.exist')
      cy.contains(
        'Table-first asset management for uploaded photos, unlinked evidence, and assignment follow-up.'
      ).should('not.exist')
      cy.contains(/assets? in (All|Unlinked) view/).should('not.exist')
      cy.contains(/Showing \d+ of \d+ media assets/).should('not.exist')
      cy.get('[role="tab"]').should('not.exist')
      cy.get('[data-testid="media-table-summary"]').should('not.exist')
    })
    cy.get('[data-testid="media-shared-table"]')
      .should('be.visible')
      .and('have.attr', 'data-table-surface', 'true')
    cy.get('[data-testid="media-table-toolbar"]').within(() => {
      cy.get('[data-testid="media-table-search-input"]').should('be.visible')
      cy.get('[data-testid="media-linkage-filter-trigger"]')
        .should('be.visible')
        .and('contain', 'All')
      cy.get('[data-testid="media-upload-action"]').should('be.enabled')
      cy.get('[data-testid="media-download-selected-action"]').should(
        'be.disabled'
      )
      cy.get('[data-testid="media-view-mode-cards"]').should('be.visible')
      cy.get('[data-testid="media-view-mode-rows"]').should('be.visible')
      cy.get('[data-testid="data-table-view-options-trigger"]').should(
        'contain',
        'View'
      )
    })
    cy.get('[data-testid="media-row-table"]')
      .should('be.visible')
      .find('tr[data-testid^="media-row-media-"]')
      .should('have.length', 3)
    cy.get('[data-testid="media-card-grid"]').should('not.exist')

    cy.get('[data-testid="media-row-media-slot-car-front"]')
      .should('contain', 'AFX Mustang front view')
      .and('contain', 'Unlinked')
      .and('contain', 'Analysis ready')
      .and('contain', 'slot-car-front-media-sl.jpg')
    cy.get('[data-testid="media-row-open-media-slot-car-front"]').should(
      'be.enabled'
    )
    cy.get('[data-testid="media-row-analyze-media-slot-car-front"]').should(
      'be.disabled'
    )
    cy.get('[data-testid="media-row-assign-media-slot-car-front"]').should(
      'be.enabled'
    )
    cy.contains(/^Ready for review$/).should('not.exist')
    cy.get('[data-testid="media-upload-action"]').and(
      'have.attr',
      'aria-label',
      'Add new asset'
    )

    chooseMediaLinkageFilter('unlinked')
    cy.wait('@unlinkedMediaAssets')
    cy.get('[data-testid="media-row-table"]')
      .find('tr[data-testid^="media-row-media-"]')
      .should('have.length', 1)
    cy.get('[data-testid="media-row-media-slot-car-front"]').should(
      'be.visible'
    )
    cy.get('[data-testid="media-row-media-porsche-box"]').should('not.exist')
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

    cy.get('[data-testid="media-row-select-media-slot-car-front"]').click()
    cy.get('[data-testid="media-download-selected-action"]')
      .should('be.enabled')
      .click()
    cy.wait('@downloadPreview')
    cy.get('[data-testid="media-download-preview"]')
      .should('be.visible')
      .and('contain', '1 file ready')
      .and('contain', 'slot-car-front-media-sl.jpg')
  })

  it('UI-SCREEN-MEDIA-003 wires Media quick actions to analysis and assignment UI', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')
    cy.intercept('POST', '/api/chat/workflow-runs', (req) => {
      expect(req.body).to.include({
        workflow_id: 'openai-image-analyze',
        capability_id: 'image_analyze',
        source_channel: 'media_workspace',
        confirmation_state: 'not_required',
      })
      expect(req.body.profile_id).to.be.a('string').and.not.equal('')
      expect(req.body.input).to.deep.equal({
        media_id: 'media-wishlist-reference',
        analysis_goal: 'identify visible item details',
      })
      expect(req.body.provider_trace).to.include({
        provider: 'openai',
        setup_needed: 'provider_test_required',
        media_access: 'read',
      })
      req.reply({
        statusCode: 201,
        body: {
          id: 'workflow-run-media-analysis-1',
          status: 'queued',
          capability_id: 'image_analyze',
          provider_trace: {
            provider: 'openai',
            setup_needed: 'provider_test_required',
            media_access: 'read',
          },
        },
      })
    }).as('analysisWorkflow')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-row-analyze-media-wishlist-reference"]')
      .should('be.enabled')
      .and('have.attr', 'aria-label', 'Analyze Wanted chassis reference')
      .click()
    cy.wait('@analysisWorkflow')
    cy.get('[data-testid="media-analysis-dialog"]')
      .should('be.visible')
      .and('contain', 'Wanted chassis reference')
    cy.get('[data-testid="media-analysis-result"]')
      .should('contain', 'Analysis workflow queued')
      .and('contain', 'workflow-run-media-analysis-1')
      .and('contain', 'media-wishlist-reference')
    cy.get('[data-testid="media-row-assign-media-wishlist-reference"]')
      .should('be.disabled')
      .and('have.attr', 'aria-label', 'Assign Wanted chassis reference')
    cy.get('[data-testid="media-analysis-dialog"]')
      .contains('button', 'Close')
      .click()

    cy.get('[data-testid="media-row-assign-media-slot-car-front"]')
      .should('be.enabled')
      .and('have.attr', 'aria-label', 'Assign AFX Mustang front view')
      .click()
    cy.get('[data-testid="media-assignment-dialog"]')
      .should('be.visible')
      .and('contain', 'AFX Mustang front view')
  })

  it('UI-SCREEN-MEDIA-012 keeps Media cards compact and responsive', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')

    cy.viewport(1280, 900)
    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')
    cy.get('[data-testid="media-view-mode-cards"]').click()

    cy.get('[data-testid="media-card-grid"]').should('be.visible')
    cy.get('[data-testid="media-card-media-slot-car-front"]').then(($card) => {
      const rect = $card[0].getBoundingClientRect()
      expect(rect.width).to.be.lessThan(240)
      expect(rect.height).to.be.lessThan(400)
    })
    cy.get('[data-testid="media-card-media-slot-car-front"]')
      .should('contain', 'AFX Mustang front view')
      .and('contain', 'slot-car-front-media-sl.jpg')
    cy.get('[data-testid="media-open-media-slot-car-front"]').should(
      'be.visible'
    )
    cy.get('[data-testid="media-analyze-media-slot-car-front"]').should(
      'be.visible'
    )
    cy.get('[data-testid="media-assign-media-slot-car-front"]').should(
      'be.visible'
    )
    cy.get('[data-testid="media-archive-media-slot-car-front"]').should(
      'be.visible'
    )

    cy.viewport(390, 844)
    cy.reload()
    cy.wait('@mediaAssets')
    cy.get('[data-testid="media-view-mode-cards"]').click()
    cy.get('[data-testid="media-card-grid"]').then(($grid) => {
      const rect = $grid[0].getBoundingClientRect()
      expect(rect.left).to.be.at.least(0)
      expect(rect.right).to.be.at.most(390)
    })
    cy.get('[data-testid="media-card-media-slot-car-front"]').then(($card) => {
      const rect = $card[0].getBoundingClientRect()
      expect(rect.width).to.be.greaterThan(300)
      expect(rect.right).to.be.at.most(390)
    })
    cy.get('[data-testid="media-card-media-slot-car-front"]')
      .should('contain', 'AFX Mustang front view')
      .and('contain', 'slot-car-front-media-sl.jpg')
    cy.get('[data-testid="media-archive-media-slot-car-front"]').should(
      'be.visible'
    )
  })

  it('UI-SCREEN-MEDIA-015 defaults to the shared Media table and keeps Cards available', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-view-mode-rows"]')
      .should('have.attr', 'aria-pressed', 'true')
      .and('contain', 'Rows')
    cy.get('[data-testid="media-shared-table"]')
      .should('be.visible')
      .and('have.attr', 'data-table-surface', 'true')
    cy.get('[data-testid="media-table-search-input"]')
      .should('be.visible')
      .type('porsche')
    cy.get('[data-testid="media-table-toolbar"]').within(() => {
      cy.get('[data-testid="media-linkage-filter-trigger"]')
        .should('be.visible')
        .and('contain', 'All')
      cy.get('[data-testid="media-view-mode-cards"]').should('be.visible')
      cy.get('[data-testid="media-view-mode-rows"]').should('be.visible')
      cy.get('[data-testid="data-table-view-options-trigger"]').should(
        'contain',
        'View'
      )
    })
    cy.get('[data-testid="media-row-table"]')
      .find('tr[data-testid^="media-row-media-"]')
      .should('have.length', 1)
    cy.get('[data-testid="media-row-media-porsche-box"]')
      .should('contain', 'Porsche 917 box side')
      .and('contain', 'Inventory linked')
    cy.get('[data-testid="media-table-search-input"]').clear()
    cy.get('[data-testid="media-row-table"]')
      .find('tr[data-testid^="media-row-media-"]')
      .should('have.length', 3)

    cy.window()
      .its('localStorage')
      .invoke('getItem', 'cabinet.viewMode.media')
      .should('eq', 'rows')

    cy.get('[data-testid="media-view-mode-cards"]').click()
    cy.window()
      .its('localStorage')
      .invoke('getItem', 'cabinet.viewMode.media')
      .should('eq', 'cards')
    cy.get('[data-testid="media-card-grid"]').should('be.visible')
    cy.get('[data-testid="media-shared-table"]')
      .should('be.visible')
      .and('have.attr', 'data-media-view-mode', 'cards')

    cy.reload()
    cy.wait('@mediaAssets')
    cy.get('[data-testid="media-view-mode-cards"]').should(
      'have.attr',
      'aria-pressed',
      'true'
    )
    cy.get('[data-testid="media-card-grid"]').should('be.visible')

    cy.get('[data-testid="media-view-mode-rows"]').click()
    cy.window()
      .its('localStorage')
      .invoke('getItem', 'cabinet.viewMode.media')
      .should('eq', 'rows')
    cy.get('[data-testid="media-shared-table"]').should('be.visible')
    cy.get('[data-testid="media-card-grid"]').should('not.exist')
  })

  it('UI-SCREEN-MEDIA-017 keeps Cards and Rows on the shared table behavior', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-workspace"]').should(
      'have.attr',
      'data-media-table-layout',
      'full-height'
    )
    cy.get('[data-testid="media-table-scroll-body"]').should('be.visible')
    cy.get('[data-testid="media-table-pagination"]').should('be.visible')

    cy.get('[data-testid="media-view-mode-cards"]').click()
    cy.get('[data-testid="media-shared-table"]')
      .should('be.visible')
      .and('have.attr', 'data-table-surface', 'true')
      .and('have.attr', 'data-media-view-mode', 'cards')
    cy.get('[data-testid="media-table-search-input"]')
      .should('be.visible')
      .clear()
      .type('porsche')
    cy.get('[data-testid="media-card-grid"]')
      .find('[data-testid^="media-card-media-"]')
      .should('have.length', 1)
    cy.get('[data-testid="media-card-media-porsche-box"]')
      .should('contain', 'Porsche 917 box side')
      .and('contain', 'Inventory linked')
    cy.get('[data-testid="media-table-pagination"]').should('be.visible')

    cy.get('[data-testid="media-view-mode-rows"]').click()
    cy.get('[data-testid="media-shared-table"]')
      .should('have.attr', 'data-media-view-mode', 'rows')
    cy.get('[data-testid="media-row-table"]')
      .find('tr[data-testid^="media-row-media-"]')
      .should('have.length', 1)
    cy.get('[data-testid="media-row-media-porsche-box"]').should(
      'contain',
      'Porsche 917 box side'
    )
  })

  it('UI-SCREEN-MEDIA-016 opens double-click metadata modal and saves edits', () => {
    let saved = false
    cy.intercept('GET', '/api/media/assets', (req) => {
      req.reply({
        statusCode: 200,
        body: saved
          ? {
              ...mediaResponse,
              assets: mediaResponse.assets.map((asset) =>
                asset.id === 'media-slot-car-front'
                  ? {
                      ...asset,
                      title: 'AFX Mustang hero angle',
                      filename: 'slot-car-hero.jpg',
                      source: 'Bench edit',
                      notes: 'Updated crop and metadata',
                      download_filename: 'afx-mustang-hero-angle.jpg',
                    }
                  : asset
              ),
            }
          : mediaResponse,
      })
    }).as('mediaAssets')
    cy.intercept(
      'PATCH',
      '/api/media/assets/media-slot-car-front/metadata',
      (req) => {
        expect(req.body).to.deep.equal({
          title: 'AFX Mustang hero angle',
          filename: 'slot-car-hero.jpg',
          source: 'Bench edit',
          download_filename: 'afx-mustang-hero-angle.jpg',
          notes: 'Updated crop and metadata',
        })
        saved = true
        req.reply({
          statusCode: 200,
          body: {
            asset_id: 'media-slot-car-front',
            title: 'AFX Mustang hero angle',
            filename: 'slot-car-hero.jpg',
            source: 'Bench edit',
            download_filename: 'afx-mustang-hero-angle.jpg',
            notes: 'Updated crop and metadata',
          },
        })
      }
    ).as('updateMediaMetadata')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-row-media-slot-car-front"]').dblclick()
    cy.get('[data-testid="media-edit-dialog"]')
      .should('be.visible')
      .and('contain', 'AFX Mustang front view')
      .and('contain', 'Original')
      .and('contain', 'Thumbnail')
      .and('contain', 'Review crop')
    cy.get('[data-testid="media-edit-thumbnail"]').should('be.visible')
    cy.get('[data-testid="media-edit-title"]')
      .should('have.value', 'AFX Mustang front view')
      .clear()
      .type('AFX Mustang hero angle')
    cy.get('[data-testid="media-edit-filename"]')
      .should('have.value', 'slot-car-front.jpg')
      .clear()
      .type('slot-car-hero.jpg')
    cy.get('[data-testid="media-edit-source"]')
      .should('have.value', 'Chat attachment')
      .clear()
      .type('Bench edit')
    cy.get('[data-testid="media-edit-download-filename"]')
      .should('have.value', 'slot-car-front-media-sl.jpg')
      .clear()
      .type('afx-mustang-hero-angle.jpg')
    cy.get('[data-testid="media-edit-notes"]')
      .should('have.value', 'Initial intake note')
      .clear()
      .type('Updated crop and metadata')
    cy.get('[data-testid="media-edit-save-action"]').click()
    cy.wait('@updateMediaMetadata')
    cy.wait('@mediaAssets')
    cy.get('[data-testid="media-edit-dialog"]').should('not.exist')
    cy.get('[data-testid="media-assignment-success"]')
      .should('be.visible')
      .and('contain', 'Media metadata updated')
    cy.get('[data-testid="media-row-media-slot-car-front"]')
      .should('contain', 'AFX Mustang hero angle')
      .and('contain', 'afx-mustang-hero-angle.jpg')
  })

  it('UI-SCREEN-MEDIA-014 supports page-wide image drop and metadata save', () => {
    let saved = false
    const createdAsset = {
      id: 'media-page-drop-created',
      title: 'Loose chassis reference',
      filename: 'loose-chassis.jpg',
      uploaded_at: '2026-06-08 16:35',
      linkage_state: 'unlinked',
      analysis_status: 'pending',
      source: 'Chat attachment',
      download_filename: 'loose-chassis-jpg-media-pa.jpg',
    }

    cy.intercept('GET', '/api/media/assets', (req) => {
      req.reply({
        statusCode: 200,
        body: saved
          ? {
              ...mediaResponse,
              summary: {
                ...mediaResponse.summary,
                total: mediaResponse.summary.total + 1,
                unlinked: mediaResponse.summary.unlinked + 1,
              },
              assets: [createdAsset, ...mediaResponse.assets],
            }
          : mediaResponse,
      })
    }).as('mediaAssets')
    cy.intercept('POST', '/api/media/assets', (req) => {
      expect(req.headers['content-type']).to.contain('multipart/form-data')
      saved = true
      req.reply({
        statusCode: 201,
        body: {
          asset_id: 'media-page-drop-created',
          filename: 'loose-chassis.jpg',
          title: 'Loose chassis reference',
          source: 'Bench intake',
          notes: 'Rear axle detail',
        },
      })
    }).as('createMediaAsset')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')
    cy.get('[data-testid="media-view-mode-cards"]').click()

    cy.get('[data-testid="media-workspace"]').selectFile(
      {
        contents: Cypress.Buffer.from('fake-image-data'),
        fileName: 'loose-chassis.jpg',
        mimeType: 'image/jpeg',
      },
      { action: 'drag-drop' }
    )
    cy.get('[data-testid="media-add-dialog"]').should('be.visible')
    cy.get('[data-testid="media-add-dropzone"]').should(
      'contain',
      'loose-chassis.jpg'
    )
    cy.get('[data-testid="media-add-title"]')
      .should('have.value', 'loose-chassis')
      .clear()
      .type('Loose chassis reference')
    cy.get('[data-testid="media-add-source"]').type('Bench intake')
    cy.get('[data-testid="media-add-notes"]').type('Rear axle detail')
    cy.get('[data-testid="media-add-save-action"]').click()
    cy.wait('@createMediaAsset')
    cy.wait('@mediaAssets')
    cy.get('[data-testid="media-add-dialog"]').should('not.exist')
    cy.get('[data-testid="media-assignment-success"]')
      .should('be.visible')
      .and('contain', 'Media asset added')
    cy.get('[data-testid="media-card-media-page-drop-created"]')
      .should('be.visible')
      .and('contain', 'Loose chassis reference')
      .and('contain', 'Unlinked')
  })

  it('UI-SCREEN-MEDIA-014 rejects unsupported files and preserves metadata', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')
    cy.intercept('POST', '/api/media/assets').as('createMediaAsset')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-upload-action"]').click()
    cy.get('[data-testid="media-add-dialog"]').should('be.visible')
    cy.get('[data-testid="media-add-title"]').type('Metadata survives')
    cy.get('[data-testid="media-add-source"]').type('Owner note')
    cy.get('[data-testid="media-add-dropzone"]').selectFile(
      {
        contents: Cypress.Buffer.from('plain-text'),
        fileName: 'not-an-image.txt',
        mimeType: 'text/plain',
      },
      { action: 'drag-drop' }
    )
    cy.get('[data-testid="media-add-error"]')
      .should('be.visible')
      .and('contain', 'Unsupported file type')
    cy.get('[data-testid="media-add-title"]').should(
      'have.value',
      'Metadata survives'
    )
    cy.get('[data-testid="media-add-source"]').should('have.value', 'Owner note')
    cy.get('@createMediaAsset.all').should('have.length', 0)
  })

  it('UI-SCREEN-MEDIA-014 preserves metadata when save fails', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')
    cy.intercept('POST', '/api/media/assets', {
      statusCode: 500,
      body: { error: 'failed_to_save_media_asset' },
    }).as('createMediaAsset')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-upload-action"]').click()
    cy.get('[data-testid="media-add-file-input"]').selectFile(
      {
        contents: Cypress.Buffer.from('fake-image-data'),
        fileName: 'failed-save.png',
        mimeType: 'image/png',
      },
      { force: true }
    )
    cy.get('[data-testid="media-add-title"]').type('Failure metadata')
    cy.get('[data-testid="media-add-source"]').type('Page button')
    cy.get('[data-testid="media-add-notes"]').type('Keep this note')
    cy.get('[data-testid="media-add-save-action"]').click()
    cy.wait('@createMediaAsset')
    cy.get('[data-testid="media-add-dialog"]').should('be.visible')
    cy.get('[data-testid="media-add-error"]')
      .should('be.visible')
      .and('contain', 'media_asset_save_500')
    cy.get('[data-testid="media-add-dropzone"]').should(
      'contain',
      'failed-save.png'
    )
    cy.get('[data-testid="media-add-title"]').should(
      'have.value',
      'Failure metadata'
    )
    cy.get('[data-testid="media-add-source"]').should('have.value', 'Page button')
    cy.get('[data-testid="media-add-notes"]').should(
      'have.value',
      'Keep this note'
    )
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

    cy.get('[data-testid="media-row-assign-media-slot-car-front"]').click()
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
    cy.get('[data-testid="media-row-media-slot-car-front"]')
      .should('contain', 'Wishlist linked')
      .and('not.contain', 'Unlinked')
    cy.get('[data-testid="media-row-assign-media-slot-car-front"]').should(
      'be.disabled'
    )
  })

  it('UI-SCREEN-MEDIA-011 keeps linkage state unchanged when assignment preview or apply fails', () => {
    cy.intercept('GET', '/api/media/assets', {
      statusCode: 200,
      body: mediaResponse,
    }).as('mediaAssets')
    cy.intercept('POST', '/api/media/assignments/preview', (req) => {
      expect(req.body).to.deep.equal({
        asset_id: 'media-slot-car-front',
        target_type: 'wishlist',
        target_id: 'missing-wishlist',
      })
      req.reply({
        statusCode: 404,
        body: { error: 'wishlist_target_not_found' },
      })
    }).as('assignmentPreviewFailed')
    cy.intercept('POST', '/api/media/assignments', (req) => {
      expect(req.body).to.deep.equal({
        asset_id: 'media-slot-car-front',
        target_type: 'wishlist',
        target_id: 'wish-slot-car',
      })
      req.reply({
        statusCode: 409,
        body: { error: 'media_assignment_conflict' },
      })
    }).as('assignmentApplyFailed')

    openMediaWorkspace()
    cy.visit('/media/')
    cy.wait('@mediaAssets')

    cy.get('[data-testid="media-row-assign-media-slot-car-front"]').click()
    cy.get('[data-testid="media-assignment-dialog"]').should('be.visible')
    cy.get('[data-testid="media-assignment-target-id"]').type(
      'missing-wishlist'
    )
    cy.get('[data-testid="media-assignment-preview-action"]').click()
    cy.wait('@assignmentPreviewFailed')
    cy.get('[data-testid="media-assignment-error"]')
      .should('be.visible')
      .and('contain', 'media_assignment_preview_404')
    cy.get('[data-testid="media-assignment-confirm-action"]').should(
      'be.disabled'
    )
    cy.get('[data-testid="media-row-media-slot-car-front"]')
      .should('contain', 'Unlinked')
      .and('not.contain', 'Wishlist linked')

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
    }).as('assignmentPreviewAllowed')

    cy.get('[data-testid="media-assignment-target-id"]')
      .clear()
      .type('wish-slot-car')
    cy.get('[data-testid="media-assignment-preview-action"]').click()
    cy.wait('@assignmentPreviewAllowed')
    cy.get('[data-testid="media-assignment-confirm-action"]')
      .should('be.enabled')
      .click()
    cy.wait('@assignmentApplyFailed')
    cy.get('[data-testid="media-assignment-dialog"]').should('be.visible')
    cy.get('[data-testid="media-assignment-error"]')
      .should('be.visible')
      .and('contain', 'media_assignment_409')
    cy.get('[data-testid="media-row-media-slot-car-front"]')
      .should('contain', 'Unlinked')
      .and('not.contain', 'Wishlist linked')
    cy.get('@mediaAssets.all').should('have.length', 1)
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
