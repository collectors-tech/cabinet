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
    cy.get('[data-testid="purchase-inbox-header-title"]').should(
      'contain',
      'Purchases'
    )
    cy.get('[data-testid="purchases-table-shell"]')
      .should('be.visible')
      .and('contain', 'Purchase')
      .and('contain', 'Source')
      .and('contain', 'Tracking')
    cy.get('[data-testid="purchases-add-button"]').click()
    cy.get('[data-testid="purchases-add-dialog"]').should('be.visible')
    cy.get('[data-testid="purchases-add-tab-new"]').should('contain', 'New')
    cy.get('[data-testid="purchases-add-tab-csv"]').should('contain', 'CSV')
    cy.get('[data-testid="purchases-add-tab-email"]').should('contain', 'Email')
    cy.contains('button', 'Cancel').click()
    cy.get('[data-testid="purchase-inbox-empty-state"]').should('be.visible')
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchases-table-row"]')
      .should('contain', 'Accompanying Flute TWM 142')
      .and('contain', 'eBay')
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

  it('COMMERCE-RECONCILIATION-007 filters Purchases rows and marks review state actions', () => {
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
                  seller_username: 'seller-one',
                },
                suggested_actions: [
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
                  seller_username: 'seller-two',
                },
                missing_fields: ['quantity', 'item_price'],
              },
            ],
          },
        ],
      },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchases-filter-result"]').should(
      'contain',
      'Showing 2 of 2 purchases'
    )

    cy.get('[data-testid="purchases-table-search"]').type('Mystery')
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Mystery purchase')
      .and('contain', 'eBay / seller-two')
    cy.get('[data-testid="purchases-filter-result"]').should(
      'contain',
      'Showing 1 of 2 purchases'
    )

    cy.get('[data-testid="purchases-table-search"]').clear()
    cy.get('[data-testid="purchases-status-filter-ready_to_link_or_convert"]')
      .click()
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Accompanying Flute TWM 142')
      .and('contain', 'ready to link or convert')
    cy.get('[data-testid="purchases-row-favorite"]')
      .click()
      .should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="purchases-row-arrived"]')
      .click()
      .should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="purchases-table-row"]').should('contain', 'Arrived')
    cy.get('[data-testid="purchases-row-rating"]')
      .click()
      .should('contain', 'Rating 4')

    cy.get('[data-testid="purchases-status-filter-needs_review"]').click()
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Mystery purchase')
    cy.get('[data-testid="purchases-table-search"]').type('no-match')
    cy.get('[data-testid="purchases-table-filter-empty-row"]').should(
      'contain',
      'No purchases match the current table filters.'
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
      .and('contain', 'Purchases could not load reviews.')
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
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')
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
    cy.get('[data-testid="forwarder-package-inbox"]')
      .should('contain', 'Purchase Source Matches')
      .and('contain', 'source-backed purchase candidates')
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
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')

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

  it('INTEGRATION-036/037 imports email notices and shows package detail provenance', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/forwarding/packages/import-email', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_email_import',
        package: {
          id: 'fwdpkg_email_001',
          profile_id: 'e2e-profile-001',
          provider: 'stackry',
          source: 'email',
          external_package_id: 'STK-EMAIL-3001',
          shipment_id: 'SHIP-3001',
          tracking_number: '1ZEMAIL3001',
          status: 'received',
          received_at: '2026-05-25T05:00:00Z',
          sender: 'Stackry Intake',
          warehouse_location: 'Locker E-5',
          weight_grams: 640,
          provenance_key: 'stackry:email:STK-EMAIL-3001',
          raw_payload: {
            message_id: 'manual-email-notice',
            sender: 'Stackry Intake',
          },
          created_at: '2026-05-25T05:01:00Z',
          updated_at: '2026-05-25T05:02:00Z',
        },
      },
    }).as('importForwarderPackageEmail')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_email_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-EMAIL-3001',
            shipment_id: 'SHIP-3001',
            tracking_number: '1ZEMAIL3001',
            status: 'received',
            received_at: '2026-05-25T05:00:00Z',
            sender: 'Stackry Intake',
            warehouse_location: 'Locker E-5',
            weight_grams: 640,
            provenance_key: 'stackry:email:STK-EMAIL-3001',
            raw_payload: {
              message_id: 'manual-email-notice',
              sender: 'Stackry Intake',
            },
            created_at: '2026-05-25T05:01:00Z',
            updated_at: '2026-05-25T05:02:00Z',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-email"]').clear()
    cy.get('[data-testid="forwarder-package-email"]').type(
      'Package ID: STK-EMAIL-3001\nStatus: Received\nShipment ID: SHIP-3001\nTracking Number: 1ZEMAIL3001\nWarehouse Location: Locker E-5\nWeight Grams: 640\nSender: Stackry Intake',
      { delay: 0 }
    )
    cy.get('[data-testid="forwarder-package-import-email"]').click()
    cy.wait('@importForwarderPackageEmail')
      .its('request.body')
      .should('include', {
        profile_id: 'e2e-profile-001',
        provider: 'stackry',
        message_id: 'manual-email-notice',
      })
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-result"]').should(
      'contain',
      'Imported email package STK-EMAIL-3001'
    )
    cy.get('[data-testid="forwarder-package-list"]')
      .should('contain', 'STK-EMAIL-3001')
      .and('contain', 'stackry / email')
      .and('contain', 'Stackry Intake')
      .and('contain', 'stackry:email:STK-EMAIL-3001')
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-detail"]')
      .should('contain', 'Shipment ID')
      .and('contain', 'SHIP-3001')
      .and('contain', 'Tracking number')
      .and('contain', '1ZEMAIL3001')
      .and('contain', 'Received')
      .and('contain', '2026-05-25T05:00:00Z')
      .and('contain', 'Created')
      .and('contain', '2026-05-25T05:01:00Z')
      .and('contain', 'Updated')
      .and('contain', '2026-05-25T05:02:00Z')
    cy.get('[data-testid="forwarder-package-raw-payload"]')
      .should('contain', 'manual-email-notice')
      .and('contain', 'Stackry Intake')

    cy.intercept('POST', '/api/forwarding/packages/import-email', {
      statusCode: 400,
      body: {
        error: 'invalid_forwarder_package_email',
        message: 'external_package_id is required',
      },
    }).as('importInvalidForwarderPackageEmail')
    cy.get('[data-testid="forwarder-package-email"]').clear()
    cy.get('[data-testid="forwarder-package-email"]').type(
      'Status: Received\nTracking Number: 1ZEMAIL3002',
      { delay: 0 }
    )
    cy.get('[data-testid="forwarder-package-import-email"]').click()
    cy.wait('@importInvalidForwarderPackageEmail')
    cy.get('[data-testid="forwarder-package-error"]')
      .should('be.visible')
      .and('contain', 'external_package_id is required')
  })

  it('INTEGRATION-039 links forwarder packages to purchase arrivals from the inbox UI', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_link_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-LINK-4001',
            shipment_id: 'SHIP-4001',
            tracking_number: '1ZLINK4001',
            status: 'received',
            sender: 'Stackry Intake',
            warehouse_location: 'Locker L-4',
            weight_grams: 720,
            provenance_key: 'stackry:email:STK-LINK-4001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')
    cy.intercept('POST', '/api/forwarding/package-links', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_reconciliation_link',
        link: {
          id: 'fwdpkg_link_001:item-expected-001:arrival-expected-001',
          profile_id: 'e2e-profile-001',
          package_id: 'fwdpkg_link_001',
          item_id: 'item-expected-001',
          lifecycle_entry_id: 'life-entry-001',
          expected_arrival_id: 'arrival-expected-001',
          source: 'manual_review',
          notes: 'Matched from package inbox review',
        },
      },
    }).as('saveForwarderPackageLink')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
      .its('request.url')
      .should('include', 'package_id=fwdpkg_link_001')
    cy.get('[data-testid="forwarder-package-link-panel"]')
      .should('be.visible')
      .and('contain', 'No reconciliation link recorded')
    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@saveForwarderPackageLink')
      .its('request.body')
      .should('include', {
        package_id: 'fwdpkg_link_001',
        item_id: 'item-expected-001',
        lifecycle_entry_id: 'life-entry-001',
        expected_arrival_id: 'arrival-expected-001',
        source: 'manual_review',
      })
    cy.get('[data-testid="forwarder-package-link-result"]')
      .should('be.visible')
      .and('contain', 'Confirmed link to item item-expected-001')

    cy.intercept('POST', '/api/forwarding/package-links', {
      statusCode: 400,
      body: {
        error: 'invalid_forwarder_package_link',
        message: 'forwarder package already linked to a different target',
      },
    }).as('rejectAmbiguousForwarderPackageLink')
    cy.get('[data-testid="forwarder-package-link-arrival"]')
      .clear()
      .type('arrival-other-002')
    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@rejectAmbiguousForwarderPackageLink')
    cy.get('[data-testid="forwarder-package-link-error"]')
      .should('be.visible')
      .and('contain', 'already linked to a different target')
  })

  it('INTEGRATION-042 confirms overrides unlinks and shows forwarder package link audit events', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_audit_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-AUDIT-5001',
            shipment_id: 'SHIP-5001',
            tracking_number: '1ZAUDIT5001',
            status: 'received',
            sender: 'Stackry Intake',
            warehouse_location: 'Locker A-5',
            weight_grams: 840,
            provenance_key: 'stackry:email:STK-AUDIT-5001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')

    let linkListRequest = 0
    cy.intercept('GET', '/api/forwarding/package-links*', (req) => {
      linkListRequest += 1
      const states = [
        { links: [], events: [] },
        {
          links: [
            {
              id: 'fwdpkg_audit_001:item-expected-001:arrival-expected-001',
              profile_id: 'e2e-profile-001',
              package_id: 'fwdpkg_audit_001',
              item_id: 'item-expected-001',
              lifecycle_entry_id: 'life-entry-001',
              expected_arrival_id: 'arrival-expected-001',
              source: 'manual_review',
              decision: 'confirmed',
              notes: 'Matched from package inbox review',
              audit_trail: [
                'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
              ],
            },
          ],
          events: [
            {
              id: 'event-confirmed',
              package_id: 'fwdpkg_audit_001',
              action: 'confirmed',
              item_id: 'item-expected-001',
              lifecycle_entry_id: 'life-entry-001',
              expected_arrival_id: 'arrival-expected-001',
              source: 'manual_review',
              notes: 'Matched from package inbox review',
              audit_trail: [
                'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
              ],
              created_at: '2026-05-25T09:00:00Z',
            },
          ],
        },
        {
          links: [
            {
              id: 'fwdpkg_audit_001:item-override-002:arrival-override-002',
              profile_id: 'e2e-profile-001',
              package_id: 'fwdpkg_audit_001',
              item_id: 'item-override-002',
              lifecycle_entry_id: 'life-entry-override-002',
              expected_arrival_id: 'arrival-override-002',
              source: 'manual_override',
              decision: 'overridden',
              notes: 'Override to corrected package target',
              audit_trail: [
                'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
            },
          ],
          events: [
            {
              id: 'event-overridden',
              package_id: 'fwdpkg_audit_001',
              action: 'overridden',
              item_id: 'item-override-002',
              lifecycle_entry_id: 'life-entry-override-002',
              expected_arrival_id: 'arrival-override-002',
              previous_item_id: 'item-expected-001',
              previous_lifecycle_entry_id: 'life-entry-001',
              previous_expected_arrival_id: 'arrival-expected-001',
              source: 'manual_override',
              notes: 'Override to corrected package target',
              audit_trail: [
                'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
              created_at: '2026-05-25T09:05:00Z',
            },
            {
              id: 'event-confirmed',
              package_id: 'fwdpkg_audit_001',
              action: 'confirmed',
              item_id: 'item-expected-001',
              lifecycle_entry_id: 'life-entry-001',
              expected_arrival_id: 'arrival-expected-001',
              source: 'manual_review',
              notes: 'Matched from package inbox review',
              audit_trail: [
                'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
              ],
              created_at: '2026-05-25T09:00:00Z',
            },
          ],
        },
        {
          links: [],
          events: [
            {
              id: 'event-unlinked',
              package_id: 'fwdpkg_audit_001',
              action: 'unlinked',
              previous_item_id: 'item-override-002',
              previous_lifecycle_entry_id: 'life-entry-override-002',
              previous_expected_arrival_id: 'arrival-override-002',
              source: 'manual_unlink',
              notes: 'Override to corrected package target',
              audit_trail: [
                'unlinked from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
              created_at: '2026-05-25T09:08:00Z',
            },
            {
              id: 'event-overridden',
              package_id: 'fwdpkg_audit_001',
              action: 'overridden',
              item_id: 'item-override-002',
              lifecycle_entry_id: 'life-entry-override-002',
              expected_arrival_id: 'arrival-override-002',
              previous_item_id: 'item-expected-001',
              previous_lifecycle_entry_id: 'life-entry-001',
              previous_expected_arrival_id: 'arrival-expected-001',
              source: 'manual_override',
              notes: 'Override to corrected package target',
              audit_trail: [
                'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
              created_at: '2026-05-25T09:05:00Z',
            },
          ],
        },
      ]
      const state = states[Math.min(linkListRequest - 1, states.length - 1)]
      req.reply({
        statusCode: 200,
        body: {
          links: state.links,
          events: state.events,
          summary: {
            count: state.links.length,
            events: state.events.length,
          },
        },
      })
    }).as('listForwarderPackageLinks')

    cy.intercept('POST', '/api/forwarding/package-links', (req) => {
      expect(req.body).to.include({
        package_id: 'fwdpkg_audit_001',
        item_id: 'item-expected-001',
        lifecycle_entry_id: 'life-entry-001',
        expected_arrival_id: 'arrival-expected-001',
        source: 'manual_review',
        decision: 'confirmed',
        override: false,
        actor: 'reviewer',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_link',
          link: {
            id: 'fwdpkg_audit_001:item-expected-001:arrival-expected-001',
            package_id: 'fwdpkg_audit_001',
            item_id: 'item-expected-001',
            expected_arrival_id: 'arrival-expected-001',
            source: 'manual_review',
            decision: 'confirmed',
          },
        },
      })
    }).as('confirmForwarderPackageLink')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Source records')
      .and('contain', '1')
      .and('contain', 'Linked')
      .and('contain', '0')
      .and('contain', 'Unlinked')
      .and('contain', '1')
    cy.get('[data-testid="forwarder-package-review-filter-result"]').should(
      'contain',
      'Showing 1 of 1 packages'
    )
    cy.get('[data-testid="forwarder-package-review-filter-linked"]').click()
    cy.get('[data-testid="forwarder-package-filter-empty"]').should(
      'contain',
      'No source records match this review state'
    )
    cy.get('[data-testid="forwarder-package-review-filter-result"]').should(
      'contain',
      'Showing 0 of 1 packages'
    )
    cy.get('[data-testid="forwarder-package-review-filter-unlinked"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-AUDIT-5001')
    cy.get('[data-testid="forwarder-package-row-evidence"]').should(
      'contain',
      'No loaded reconciliation evidence'
    )
    cy.get('[data-testid="forwarder-package-review-filter-all"]').click()
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('be.visible')
      .and('contain', 'No link decisions recorded yet.')

    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@confirmForwarderPackageLink')
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-result"]')
      .should('be.visible')
      .and('contain', 'Confirmed link to item item-expected-001')
    cy.get('[data-testid="forwarder-package-link-state"]')
      .should('contain', 'confirmed to item item-expected-001')
      .and('contain', 'arrival arrival-expected-001')
      .and('contain', 'Source manual_review')
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Linked')
      .and('contain', '1')
      .and('contain', 'Unlinked')
      .and('contain', '0')
      .and('contain', 'Audit events')
      .and('contain', '1')
    cy.get('[data-testid="forwarder-package-review-filter-linked"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-AUDIT-5001')
    cy.get('[data-testid="forwarder-package-row-evidence"]')
      .should('contain', '1 active link')
      .and('contain', '1 audit event')
    cy.get('[data-testid="forwarder-package-review-filter-all"]').click()
    cy.get('[data-testid="forwarder-package-link-audit-trail"]').should(
      'contain',
      'confirmed from purchase inbox UI'
    )
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('contain', 'confirmed item item-expected-001')
      .and('contain', 'lifecycle life-entry-001')
      .and('contain', '2026-05-25T09:00:00Z')
      .and('contain', 'via manual_review')
    cy.get('[data-testid="forwarder-package-link-event-audit-trail"]').should(
      'contain',
      'confirmed from purchase inbox UI'
    )

    cy.intercept('POST', '/api/forwarding/package-links', (req) => {
      expect(req.body).to.include({
        package_id: 'fwdpkg_audit_001',
        item_id: 'item-override-002',
        lifecycle_entry_id: 'life-entry-override-002',
        expected_arrival_id: 'arrival-override-002',
        source: 'manual_override',
        decision: 'overridden',
        override: true,
        actor: 'reviewer',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_link',
          link: {
            id: 'fwdpkg_audit_001:item-override-002:arrival-override-002',
            package_id: 'fwdpkg_audit_001',
            item_id: 'item-override-002',
            expected_arrival_id: 'arrival-override-002',
            source: 'manual_override',
            decision: 'overridden',
          },
        },
      })
    }).as('overrideForwarderPackageLink')

    cy.get('[data-testid="forwarder-package-link-item"]')
      .clear()
      .type('item-override-002')
    cy.get('[data-testid="forwarder-package-link-lifecycle"]')
      .clear()
      .type('life-entry-override-002')
    cy.get('[data-testid="forwarder-package-link-arrival"]')
      .clear()
      .type('arrival-override-002')
    cy.get('[data-testid="forwarder-package-link-notes"]')
      .clear()
      .type('Override to corrected package target')
    cy.get('[data-testid="forwarder-package-link-override"]').click()
    cy.wait('@overrideForwarderPackageLink')
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-result"]').should(
      'contain',
      'Override linked to item item-override-002'
    )
    cy.get('[data-testid="forwarder-package-link-state"]')
      .should('contain', 'overridden to item item-override-002')
      .and('contain', 'Source manual_override')
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('contain', 'overridden item item-override-002')
      .and('contain', 'previous item item-expected-001')
      .and('contain', 'previous item item-expected-001 / lifecycle life-entry-001 / arrival arrival-expected-001')
      .and('contain', '2026-05-25T09:05:00Z')

    cy.intercept('DELETE', '/api/forwarding/package-links*', (req) => {
      expect(req.url).to.include('package_id=fwdpkg_audit_001')
      expect(req.body).to.include({
        source: 'manual_unlink',
        actor: 'reviewer',
        notes: 'Override to corrected package target',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'unlinked from purchase inbox UI: item-override-002 / arrival-override-002',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_unlink',
          event: {
            id: 'event-unlinked',
            package_id: 'fwdpkg_audit_001',
            action: 'unlinked',
            previous_item_id: 'item-override-002',
            source: 'manual_unlink',
          },
        },
      })
    }).as('unlinkForwarderPackage')

    cy.get('[data-testid="forwarder-package-link-unlink"]').click()
    cy.wait('@unlinkForwarderPackage')
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-result"]').should(
      'contain',
      'Unlinked package from reconciliation target'
    )
    cy.get('[data-testid="forwarder-package-link-empty"]').should(
      'contain',
      'No reconciliation link recorded'
    )
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Linked')
      .and('contain', '0')
      .and('contain', 'Unlinked')
      .and('contain', '1')
      .and('contain', 'Audit events')
      .and('contain', '2')
    cy.get('[data-testid="forwarder-package-review-filter-unlinked"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-AUDIT-5001')
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('contain', 'unlinked')
      .and('contain', 'previous item item-override-002')
      .and('contain', 'previous item item-override-002 / lifecycle life-entry-override-002 / arrival arrival-override-002')
      .and('contain', 'via manual_unlink')
    cy.get('[data-testid="forwarder-package-link-event-audit-trail"]').should(
      'contain',
      'unlinked from purchase inbox UI'
    )
  })

  it('INTEGRATION-043 shows forwarder package match suggestions and prepares confirmation', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_suggest_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-SUGGEST-6001',
            shipment_id: 'SHIP-6001',
            tracking_number: '1ZSUGGEST6001',
            status: 'received',
            sender: 'slot shop',
            warehouse_location: 'Locker S-6',
            weight_grams: 760,
            provenance_key: 'stackry:email:STK-SUGGEST-6001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')

    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], events: [], summary: { count: 0, events: 0 } },
    }).as('listForwarderPackageLinks')

    let scopedSuggestionRequestSeen = false
    cy.intercept('GET', '/api/forwarding/package-match-suggestions*', (req) => {
      expect(req.query).to.include({ confidence_label: 'high' })
      if (req.query.package_id === 'fwdpkg_suggest_001') {
        scopedSuggestionRequestSeen = true
      }
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_match_suggestions',
          mutable: false,
          confidence_filter: 'high',
          suggestions: [
            {
              id: 'suggestion-fwdpkg-001',
              package_id: 'fwdpkg_suggest_001',
              item_id: 'item-suggested-001',
              lifecycle_entry_id: 'life-suggested-001',
              expected_arrival_id: 'arrival-suggested-001',
              confidence_score: 94,
              confidence_label: 'high',
              signals: [
                {
                  name: 'tracking',
                  score: 40,
                  evidence: '1ZSUGGEST6001 matched purchase notes',
                },
                {
                  name: 'seller',
                  score: 20,
                  evidence: 'slot shop matched package sender',
                },
              ],
              audit_trail: [
                'suggested_match package=fwdpkg_suggest_001 item=item-suggested-001 confidence=high score=94',
              ],
            },
          ],
          summary: {
            count: 1,
            scoped_packages: 1,
            high_confidence: 1,
            medium_confidence: 0,
            low_confidence: 0,
          },
        },
      })
    }).as('listForwarderPackageMatchSuggestions')

    cy.intercept('POST', '/api/forwarding/package-links', (req) => {
      expect(req.body).to.include({
        package_id: 'fwdpkg_suggest_001',
        item_id: 'item-suggested-001',
        lifecycle_entry_id: 'life-suggested-001',
        expected_arrival_id: 'arrival-suggested-001',
        source: 'suggested_match',
        decision: 'confirmed',
        override: false,
        actor: 'reviewer',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'confirmed from purchase inbox UI: item-suggested-001 / arrival-suggested-001',
        'suggested_match package=fwdpkg_suggest_001 item=item-suggested-001 confidence=high score=94',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_link',
          link: {
            id: 'fwdpkg_suggest_001:item-suggested-001:arrival-suggested-001',
            package_id: 'fwdpkg_suggest_001',
            item_id: 'item-suggested-001',
            expected_arrival_id: 'arrival-suggested-001',
            source: 'suggested_match',
            decision: 'confirmed',
          },
        },
      })
    }).as('confirmSuggestedForwarderPackageLink')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-confidence-filter-high"]').click()
    cy.get('[data-testid="forwarder-package-match-suggestions-load"]').click()
    cy.wait('@listForwarderPackageMatchSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-result"]')
      .should('be.visible')
      .and('contain', 'Found 1 package match suggestion')
      .and('contain', 'high confidence')
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Source records')
      .and('contain', 'Suggestions')
      .and('contain', '1')
    cy.get('[data-testid="forwarder-package-suggestion-summary"]')
      .should('contain', 'Candidates')
      .and('contain', '1')
      .and('contain', 'Scoped packages')
      .and('contain', 'High confidence')
      .and('contain', 'Medium confidence')
      .and('contain', 'Low confidence')
      .and('contain', 'Active filter')
      .and('contain', 'high')
      .and('contain', '0')
    cy.get('[data-testid="forwarder-package-review-filter-suggested"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-SUGGEST-6001')
    cy.get('[data-testid="forwarder-package-row-evidence"]').should(
      'contain',
      '1 suggestion'
    )
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
    cy.get(
      '[data-testid="forwarder-package-match-suggestions-load-scoped"]'
    ).click()
    cy.wait('@listForwarderPackageMatchSuggestions')
    cy.then(() => {
      expect(scopedSuggestionRequestSeen).to.equal(true)
    })
    cy.get('[data-testid="forwarder-package-suggestion-result"]')
      .should('contain', 'Found 1 package match suggestion')
      .and('contain', 'for package fwdpkg_suggest_001')
      .and('contain', 'high confidence')
    cy.get('[data-testid="forwarder-package-match-suggestions"]')
      .should('be.visible')
      .and('contain', 'Suggested purchase matches')
      .and('contain', 'high match to item item-suggested-001')
      .and('contain', 'Score 94')
      .and('contain', 'arrival arrival-suggested-001')
    cy.get('[data-testid="forwarder-package-match-signals"]')
      .should('contain', 'tracking')
      .and('contain', '1ZSUGGEST6001 matched purchase notes')
    cy.get('[data-testid="forwarder-package-match-audit-trail"]').should(
      'contain',
      'suggested_match package=fwdpkg_suggest_001'
    )

    cy.get('[data-testid="forwarder-package-match-suggestion-use"]').click()
    cy.get('[data-testid="forwarder-package-link-item"]').should(
      'have.value',
      'item-suggested-001'
    )
    cy.get('[data-testid="forwarder-package-link-source"]').should(
      'have.value',
      'suggested_match'
    )
    cy.get('[data-testid="forwarder-package-link-result"]').should(
      'contain',
      'Prepared suggested match for item item-suggested-001'
    )
    cy.get('[data-testid="forwarder-package-link-panel"]')
      .should('contain', 'Match this source evidence')
      .and('contain', 'Confirm link')
      .and('contain', 'Override link')
      .and('contain', 'No reconciliation link recorded for this source record')
    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@confirmSuggestedForwarderPackageLink')
  })

  it('INTEGRATION-059 keeps package rows visible across empty and failed suggestion loads', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_empty_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'manual',
            external_package_id: 'STK-EMPTY-7001',
            shipment_id: 'SHIP-7001',
            tracking_number: '1ZEMPTY7001',
            status: 'received',
            warehouse_location: 'Locker E-7',
            weight_grams: 510,
            provenance_key: 'stackry:manual:STK-EMPTY-7001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], events: [], summary: { count: 0, events: 0 } },
    }).as('listForwarderPackageLinks')

    let suggestionRequestCount = 0
    cy.intercept('GET', '/api/forwarding/package-match-suggestions*', (req) => {
      suggestionRequestCount += 1
      if (suggestionRequestCount === 1) {
        req.reply({
          statusCode: 200,
          body: {
            mode: 'forwarder_package_match_suggestions',
            mutable: false,
            confidence_filter: 'medium',
            suggestions: [],
            summary: {
              count: 0,
              scoped_packages: 0,
              high_confidence: 0,
              medium_confidence: 0,
              low_confidence: 0,
            },
          },
        })
        return
      }
      if (suggestionRequestCount === 2) {
        req.reply({
          statusCode: 503,
          body: { error: 'suggestions_unavailable' },
        })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_match_suggestions',
          mutable: false,
          suggestions: [],
          summary: { count: 0, scoped_packages: 0 },
        },
      })
    }).as('loadForwarderPackageSuggestions')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-EMPTY-7001')

    cy.get('[data-testid="forwarder-package-confidence-filter-medium"]').click()
    cy.get('[data-testid="forwarder-package-match-suggestions-load"]').click()
    cy.wait('@loadForwarderPackageSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-result"]')
      .should('be.visible')
      .and('contain', 'No package match suggestions matched the current inbox')
      .and('contain', 'medium confidence')
      .and('contain', 'Package rows were left unchanged')
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-EMPTY-7001')

    cy.get('[data-testid="forwarder-package-match-suggestions-load"]').click()
    cy.wait('@loadForwarderPackageSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-error"]')
      .should('be.visible')
      .and('contain', 'Match suggestions could not load')
      .and('contain', 'forwarder_package_match_suggestions_503')
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-EMPTY-7001')

    cy.get('[data-testid="forwarder-package-suggestion-retry"]').click()
    cy.wait('@loadForwarderPackageSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-error"]').should(
      'not.exist'
    )
  })
})
