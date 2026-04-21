describe('settings/operations', () => {
  function signInToOperations() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Foperations')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/operations\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-001 renders runtime metadata and recovery visibility', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-test123',
        build_date: '2026-04-22',
        bind_mode: 'lan',
        runtime_host: '192.168.1.53',
        runtime_port: 17882,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: true,
      },
    }).as('runtimeRecovery')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'rev-test123'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      '192.168.1.53:17882'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'lan'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'stable'
    )
    cy.get('[data-testid="settings-operations-recovery-card"]').should(
      'contain',
      'Recovery required'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-002 retries runtime visibility without route reload', () => {
    let runtimeAttempt = 0
    let recoveryAttempt = 0

    cy.intercept('GET', '/api/runtime', (req) => {
      runtimeAttempt += 1
      if (runtimeAttempt === 1) {
        req.reply(503, { error: 'runtime_unavailable' })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          app_version: 'rev-recovered',
          build_date: '2026-04-22',
          bind_mode: 'loopback',
          runtime_host: '127.0.0.1',
          runtime_port: 17880,
          update_channel: 'stable',
          update_public_key_configured: false,
        },
      })
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', (req) => {
      recoveryAttempt += 1
      if (recoveryAttempt === 1) {
        req.reply(503, { error: 'runtime_recovery_unavailable' })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          recovery_required: false,
        },
      })
    }).as('runtimeRecovery')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.get('[data-testid="settings-operations-runtime-error"]').should(
      'contain',
      'Runtime information is unavailable right now.'
    )

    cy.get('[data-testid="settings-operations-retry"]').click()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-runtime-error"]').should('not.exist')
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'rev-recovered'
    )
    cy.get('[data-testid="settings-operations-recovery-card"]').should(
      'contain',
      'No recovery required'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-003 exports snapshot data and shows dry-run conflicts', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-export',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('GET', '/api/data/export/json', {
      statusCode: 200,
      body: {
        schema_version: 1,
        exported_at: '2026-04-22T12:00:00Z',
        items: [
          {
            brand: 'AFX',
            category: 'Slot',
            part_number: 'OPS-001',
            title: 'Exported Item',
            instances: [],
          },
        ],
      },
    }).as('exportJson')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 200,
      body: {
        total_items: 2,
        new_items: 1,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'OPS-001',
            existing_id: 'item-1',
          },
        ],
      },
    }).as('importDryRun')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-export-json"]').click()
    cy.wait('@exportJson')
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Exported 1 item snapshot.'
    )

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-001","title":"Conflict"},{"brand":"AFX","category":"Slot","part_number":"OPS-NEW","title":"New Item"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.get('[data-testid="settings-operations-import-summary"]').should(
      'contain',
      '2 items'
    )
    cy.get('[data-testid="settings-operations-import-summary"]').should(
      'contain',
      '1 conflict'
    )
    cy.get('[data-testid="settings-operations-import-summary"]').should(
      'contain',
      'OPS-001'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-004 reports dry-run failure without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-import-error',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 400,
      body: {
        error: 'failed_to_dry_run_import',
      },
    }).as('importDryRun')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-ERR","title":"Broken"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Import dry-run failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-005 applies a reviewed JSON import with the selected conflict action', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-import-apply',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 200,
      body: {
        total_items: 2,
        new_items: 1,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'OPS-APPLY-001',
            existing_id: 'item-apply-1',
          },
        ],
      },
    }).as('importDryRun')
    cy.intercept('POST', '/api/data/import/json/apply', (req) => {
      expect(req.body).to.deep.equal({
        snapshot: {
          schema_version: 1,
          items: [
            {
              brand: 'AFX',
              category: 'Slot',
              part_number: 'OPS-APPLY-001',
              title: 'Conflict',
            },
            {
              brand: 'AFX',
              category: 'Slot',
              part_number: 'OPS-APPLY-NEW',
              title: 'New Item',
            },
          ],
        },
        options: {
          default_action: 'create',
        },
      })
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
        },
      })
    }).as('importApply')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-APPLY-001","title":"Conflict"},{"brand":"AFX","category":"Slot","part_number":"OPS-APPLY-NEW","title":"New Item"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.get('[data-testid="settings-operations-import-default-action"]').click()
    cy.get('[role="option"]').contains('Create new item').click()
    cy.get('[data-testid="settings-operations-import-json-apply"]').click()
    cy.wait('@importApply')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Import applied successfully.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-006 reports import apply failure without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-import-apply-error',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 200,
      body: {
        total_items: 1,
        new_items: 0,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'OPS-APPLY-ERR',
            existing_id: 'item-apply-err',
          },
        ],
      },
    }).as('importDryRun')
    cy.intercept('POST', '/api/data/import/json/apply', {
      statusCode: 400,
      body: {
        error: 'failed_to_apply_import',
      },
    }).as('importApply')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-APPLY-ERR","title":"Broken"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.get('[data-testid="settings-operations-import-json-apply"]').click()
    cy.wait('@importApply')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Import apply failed.'
    )
  })
})
