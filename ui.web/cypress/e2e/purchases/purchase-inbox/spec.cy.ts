describe('purchases/purchase-inbox', () => {
  it('EBAY-PURCHASE-CAPTURE-006 reviews captured purchases before confirmed mutation actions', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/integrations/ebay/purchase-inbox/reviews', {
      statusCode: 200,
      body: {
        source: 'ebay_purchase_capture',
        profile_id: 'e2e-profile-001',
        reviews: [
          {
            status: 'needs_review',
            order: {
              order_id: '20-14595-70928',
              seller_usernames: ['seller-one'],
              order_total: 'AU $8.10',
              currency: 'AUD',
            },
            items: [
              {
                status: 'ready_to_link_or_convert',
                item: {
                  transaction_id: '10080684936020',
                  listing_title: 'Accompanying Flute listing',
                  purchased_identity: 'Accompanying Flute TWM 142',
                  quantity: 4,
                  item_price: 'AU $2.40',
                },
                suggested_actions: [
                  {
                    id: 'link_existing_inventory_item',
                    label: 'Link existing inventory item',
                    scope: 'item',
                    target_key: '10080684936020',
                    requires_confirmation: true,
                  },
                  {
                    id: 'convert_to_inventory_item',
                    label: 'Convert to inventory item',
                    scope: 'item',
                    target_key: '10080684936020',
                    requires_confirmation: true,
                  },
                ],
              },
              {
                status: 'needs_review',
                item: {
                  listing_id: '316046161179',
                  listing_title: 'Mystery purchase',
                },
                missing_fields: ['quantity', 'item_price'],
                suggested_actions: [
                  {
                    id: 'complete_purchase_item_fields',
                    label: 'Complete missing purchase item fields',
                    scope: 'item',
                    target_key: '316046161179:Mystery purchase',
                    requires_confirmation: false,
                  },
                ],
              },
            ],
          },
        ],
      },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="purchase-inbox-empty-state"]').should('be.visible')
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchase-inbox-ready-state"]')
      .should('contain', '20-14595-70928')
      .and('contain', 'Accompanying Flute TWM 142')
      .and('contain', 'Missing: quantity, item_price')
    cy.contains(
      '[data-testid="purchase-inbox-suggested-action"]',
      'Convert to inventory item'
    ).click()
    cy.get('[data-testid="purchase-inbox-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="purchase-inbox-confirm-action"]').click()
    cy.get('[data-testid="purchase-inbox-action-result"]').should(
      'contain',
      'Queued after confirmation: Convert to inventory item: 10080684936020'
    )
  })

  it('EBAY-PURCHASE-CAPTURE-006 exposes a retryable error state', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/integrations/ebay/purchase-inbox/reviews', {
      statusCode: 500,
      body: { error: 'failed' },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchase-inbox-error-state"]')
      .should('be.visible')
      .and('contain', 'Purchase Inbox could not load reviews.')
  })

  it('EBAY-PURCHASE-CAPTURE-006 exposes loading state while reviews are prepared', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/integrations/ebay/purchase-inbox/reviews', (req) => {
      req.reply({
        delay: 500,
        statusCode: 200,
        body: {
          source: 'ebay_purchase_capture',
          profile_id: 'e2e-profile-001',
          reviews: [],
        },
      })
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.get('[data-testid="purchase-inbox-loading-state"]')
      .should('be.visible')
      .and('contain', 'Preparing purchase review records...')
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchase-inbox-loading-state"]').should('not.exist')
    cy.get('[data-testid="purchase-inbox-empty-state"]').should('be.visible')
  })

  it('INTEGRATION-032 imports and lists manual forwarder packages', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_test_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'manual',
            external_package_id: 'STK-PKG-1001',
            shipment_id: 'SHIP-1001',
            tracking_number: '1Z999AA10123456784',
            status: 'received',
            warehouse_location: 'Locker A-12',
            weight_grams: 420,
            provenance_key: 'stackry:manual:STK-PKG-1001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('POST', '/api/forwarding/packages', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_upsert',
        package: {
          id: 'fwdpkg_test_001',
          profile_id: 'e2e-profile-001',
          provider: 'stackry',
          source: 'manual',
          external_package_id: 'STK-PKG-1001',
          shipment_id: 'SHIP-1001',
          tracking_number: '1Z999AA10123456784',
          status: 'received',
          warehouse_location: 'Locker A-12',
          weight_grams: 420,
          provenance_key: 'stackry:manual:STK-PKG-1001',
        },
      },
    }).as('importForwarderPackage')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-inbox"]').should('be.visible')
    cy.get('[data-testid="forwarder-package-import"]').click()
    cy.wait('@importForwarderPackage')
      .its('request.body')
      .should('include', {
        profile_id: 'e2e-profile-001',
        provider: 'stackry',
        source: 'manual',
        external_package_id: 'STK-PKG-1001',
        status: 'received',
        weight_grams: 420,
      })
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-result"]').should(
      'contain',
      'Imported package STK-PKG-1001'
    )
    cy.get('[data-testid="forwarder-package-list"]')
      .should('contain', 'STK-PKG-1001')
      .and('contain', 'stackry / manual')
      .and('contain', 'Locker A-12')
      .and('contain', 'stackry:manual:STK-PKG-1001')
  })

  it('INTEGRATION-032 shows forwarder package import validation errors', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/forwarding/packages', {
      statusCode: 400,
      body: {
        error: 'invalid_forwarder_package',
        message: 'external_package_id is required',
      },
    }).as('importForwarderPackage')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-external-id"]').clear()
    cy.get('[data-testid="forwarder-package-import"]').click()
    cy.wait('@importForwarderPackage')
    cy.get('[data-testid="forwarder-package-error"]')
      .should('be.visible')
      .and('contain', 'external_package_id is required')
  })

  it('INTEGRATION-034 imports CSV forwarder package rows and reports row errors', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/forwarding/packages/import-csv', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_csv_import',
        imported: [
          {
            id: 'fwdpkg_csv_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'csv',
            external_package_id: 'STK-CSV-2001',
            shipment_id: 'SHIP-2001',
            tracking_number: '1ZCSV2001',
            status: 'received',
            warehouse_location: 'Locker C-4',
            weight_grams: 520,
            provenance_key: 'stackry:csv:STK-CSV-2001',
          },
        ],
        errors: [{ row: 3, error: 'external_package_id is required' }],
        summary: { imported: 1, errors: 1 },
      },
    }).as('importForwarderPackageCSV')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_csv_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'csv',
            external_package_id: 'STK-CSV-2001',
            shipment_id: 'SHIP-2001',
            tracking_number: '1ZCSV2001',
            status: 'received',
            warehouse_location: 'Locker C-4',
            weight_grams: 520,
            provenance_key: 'stackry:csv:STK-CSV-2001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-csv"]').clear()
    cy.get('[data-testid="forwarder-package-csv"]').type(
      'Stackry Package ID,Status,Shipment ID,Tracking Number,Warehouse Location,Weight Grams\nSTK-CSV-2001,received,SHIP-2001,1ZCSV2001,Locker C-4,520\n,received,SHIP-2002,1ZCSV2002,Locker D-8,420',
      { delay: 0 }
    )
    cy.get('[data-testid="forwarder-package-import-csv"]').click()
    cy.wait('@importForwarderPackageCSV')
      .its('request.body')
      .should('include', {
        profile_id: 'e2e-profile-001',
        provider: 'stackry',
      })
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-result"]')
      .should('contain', 'Imported 1 CSV package')
      .and('contain', '1 row needs attention')
    cy.get('[data-testid="forwarder-package-csv-errors"]')
      .should('contain', 'Row 3')
      .and('contain', 'external_package_id is required')
    cy.get('[data-testid="forwarder-package-list"]')
      .should('contain', 'STK-CSV-2001')
      .and('contain', 'stackry / csv')
      .and('contain', 'stackry:csv:STK-CSV-2001')
  })
})
