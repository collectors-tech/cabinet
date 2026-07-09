describe('settings/agent-skills', () => {
  function signInToSkills() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fskills')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/skills\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('AGENT-SKILLS-REGISTRY-008 lists skills opens details and imports a local archive disabled by default', () => {
    const importedSkill = {
      id: 'cabinet.imported.parts_lookup',
      version: '1.0.0',
      display_name: 'Parts Lookup Import',
      description: 'Imported local skill for parts lookup.',
      category: 'inventory',
      source: 'archive',
      status: 'disabled',
      safety_level: 'preview-only',
      required_context: ['profile'],
      capabilities: ['inventory.search'],
      permissions: {
        local_read: true,
        local_write: false,
        requires_confirm: false,
      },
      audit_behavior: 'records import provenance',
      provenance: 'local-folder:C:/cabinet/skills/parts-lookup',
      built_in: false,
      removable: true,
      enabled: false,
      executable: false,
      validation_warnings: ['Installed disabled until reviewed.'],
    }
    const skills = [
      {
        id: 'cabinet.navigate.open_surface',
        version: '1.0.0',
        display_name: 'Open Cabinet surface',
        description: 'Navigate to a governed Cabinet route.',
        category: 'navigation',
        source: 'built-in',
        status: 'available',
        safety_level: 'read-only',
        required_context: ['profile', 'route'],
        capabilities: ['navigate.open_surface'],
        ui_targets: ['shell.navigation'],
        permissions: {
          local_read: true,
          requires_confirm: false,
        },
        audit_behavior: 'records navigation command',
        provenance: 'cabinet built-in',
        built_in: true,
        removable: false,
        enabled: true,
        executable: true,
      },
      {
        id: 'cabinet.guided.inventory.update_item',
        version: '1.0.0',
        display_name: 'Guided inventory update',
        description: 'Guided inventory update walkthrough.',
        category: 'inventory',
        source: 'built-in',
        status: 'requires-implementation',
        safety_level: 'confirm-required',
        required_context: ['profile', 'selected_item'],
        guided_workflows: ['guided.inventory.update_item'],
        permissions: {
          local_read: true,
          local_write: true,
          requires_confirm: true,
        },
        audit_behavior: 'records guided workflow request',
        provenance: 'cabinet built-in',
        built_in: true,
        removable: false,
        enabled: true,
        executable: false,
      },
    ]

    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/agent/skills?profile_id=e2e-profile-001', (req) => {
      req.reply({
        statusCode: 200,
        body: { profile_id: 'e2e-profile-001', skills },
      })
    }).as('loadSkills')
    cy.intercept('POST', '/api/agent/skills/import', (req) => {
      expect(req.body).to.deep.equal({
        profile_id: 'e2e-profile-001',
        source_path: 'C:/cabinet/skills/parts-lookup',
      })
      skills.push(importedSkill)
      req.reply({
        statusCode: 200,
        body: {
          profile_id: 'e2e-profile-001',
          result: {
            state: 'installed-disabled',
            skill: importedSkill,
            warnings: ['Installed disabled until reviewed.'],
            errors: [],
          },
        },
      })
    }).as('importSkill')

    signInToSkills()
    cy.wait('@activeProfile')
    cy.wait('@loadSkills')

    cy.get('[data-testid="settings-skills-page"]').should('be.visible')
    cy.get('[data-testid="settings-skills-summary-installed"]').should(
      'contain',
      '0'
    )
    cy.get('[data-testid="settings-skills-summary-enabled"]').should(
      'contain',
      '2'
    )
    cy.get('[data-testid="settings-skills-row"]').should('have.length', 2)
    cy.get('[data-testid="settings-skills-table"]')
      .should('contain', 'Open Cabinet surface')
      .and('contain', 'Guided inventory update')
      .and('contain', 'Requires Implementation')

    cy.get(
      '[data-testid="settings-skills-detail-cabinet.guided.inventory.update_item"]'
    ).click()
    cy.get('[data-testid="settings-skills-detail-panel"]')
      .should('contain', 'Guided inventory update')
      .and('contain', 'Required context')
      .and('contain', 'profile, selected_item')
      .and('contain', 'Enable / disable')
      .and('contain', 'Built-in skills stay enabled')

    cy.get('body').type('{esc}')
    cy.get('[data-testid="settings-skills-import-open"]').click()
    cy.get('[data-testid="settings-skills-import-panel"]').should(
      'contain',
      'Marketplace browsing is not available'
    )
    cy.get('[data-testid="settings-skills-import-source"]').type(
      'C:/cabinet/skills/parts-lookup'
    )
    cy.get('[data-testid="settings-skills-import-submit"]').click()
    cy.wait('@importSkill')
    cy.wait('@loadSkills')
    cy.get('[data-testid="settings-skills-import-result"]')
      .should('contain', 'Installed Disabled')
      .and('contain', 'Parts Lookup Import')
      .and('contain', 'Installed disabled until reviewed.')
    cy.get('[data-testid="settings-skills-summary-installed"]').should(
      'contain',
      '1'
    )
    cy.get('[data-testid="settings-skills-table"]')
      .should('contain', 'Parts Lookup Import')
      .and('contain', 'Archive')
      .and('contain', 'Disabled')
  })

  it('AGENT-SKILLS-REGISTRY-010 keeps marketplace unavailable and surfaces import validation failures', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/agent/skills?profile_id=e2e-profile-001', {
      statusCode: 200,
      body: { profile_id: 'e2e-profile-001', skills: [] },
    }).as('loadSkills')
    cy.intercept('POST', '/api/agent/skills/import', {
      statusCode: 400,
      body: {
        profile_id: 'e2e-profile-001',
        result: {
          state: 'blocked-invalid-manifest',
          errors: ['manifest cabinet.skill.json is required'],
        },
      },
    }).as('importSkillFailure')

    signInToSkills()
    cy.wait('@activeProfile')
    cy.wait('@loadSkills')

    cy.contains('Marketplace browsing, ratings, payments, and remote installs are not available.').should(
      'be.visible'
    )
    cy.get('[data-testid="settings-skills-import-open"]').click()
    cy.get('[data-testid="settings-skills-import-source"]').type(
      'C:/cabinet/skills/broken'
    )
    cy.get('[data-testid="settings-skills-import-submit"]').click()
    cy.wait('@importSkillFailure')
    cy.get('[data-testid="settings-skills-import-error"]')
      .should('contain', 'manifest cabinet.skill.json is required')
    cy.get('[data-testid="settings-skills-table"]').should(
      'contain',
      'No Agent Skills are available'
    )
  })
})
