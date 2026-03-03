describe('SETUP-WIZ', () => {
  function enterSetupFormMode() {
    cy.get('[data-testid="setup-start"]').click();
    cy.get('[data-testid="setup-instance-name"]').should('be.visible');
  }

  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);
    cy.intercept('GET', '/api/runtime/setup-status').as('setupStatus');
    cy.visit('/sign-in');
    cy.wait('@setupStatus');
    cy.get('[data-testid="setup-wizard"]').should('be.visible');
  });

  it('UC-SW-06 setup-wizard-progress-template shows step header, percentage, and footer actions', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 1 OF 3');
    cy.get('[data-testid="setup-step-percent"]').should('contain.text', '33%');
    cy.get('[data-testid="setup-progress-bar"]')
      .should('have.attr', 'aria-valuenow')
      .and('eq', '33');
    cy.get('[data-testid="setup-prev"]').should('be.disabled');
    cy.get('[data-testid="setup-next"]').should('be.visible').and('not.be.disabled');
    cy.get('[data-testid="setup-complete"]').should('not.exist');

    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 2 OF 3');
    cy.get('[data-testid="setup-step-percent"]').should('contain.text', '67%');
    cy.get('[data-testid="setup-prev"]').should('not.be.disabled');
  });

  it('UC-SW-12 setup-wizard-welcome-actions renders start/import actions before form fields', () => {
    cy.get('[data-testid="setup-start"]').should('be.visible');
    cy.get('[data-testid="setup-import-toggle"]').should('be.visible');
    cy.get('[data-testid="setup-instance-name"]').should('not.exist');
    cy.get('[data-testid="setup-profile-key"]').should('not.exist');

    cy.get('[data-testid="setup-start"]').click();
    cy.get('[data-testid="setup-instance-name"]').should('be.visible');
    cy.get('[data-testid="setup-profile-key"]').should('be.visible');
  });

  it('UC-SW-13 setup-wizard-import-existing-config loads external config and exits setup mode', () => {
    cy.request('POST', '/api/test/runtime/setup-import-source', {
      mode: 'valid',
    })
      .its('body')
      .then((seed) => {
        cy.get('[data-testid="setup-import-toggle"]').click();
        cy.get('[data-testid="setup-import-source-path"]')
          .clear()
          .type(seed.source_path);
        cy.get('[data-testid="setup-import-submit"]').click();
      });

    cy.get('[data-testid="setup-wizard"]').should('not.exist');
    cy.contains('Sign in').should('be.visible');

    cy.request('GET', '/api/runtime/setup-status').its('body.setup_required').should('eq', false);
  });

  it('UC-SW-14 setup-wizard-identity-path-preview shows config destination path', () => {
    cy.request('GET', '/api/runtime/setup-status').then((statusResponse) => {
      const expectedPath = statusResponse.body.config_path as string;
      enterSetupFormMode();
      cy.get('[data-testid="setup-config-path-preview"]')
        .should('be.visible')
        .and('contain.text', expectedPath);
    });
  });

  it('UC-SW-15 setup-wizard-optional-profile-key auto-derives profile key from instance name', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('My Fancy Instance');
    cy.get('[data-testid="setup-profile-key"]').clear();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.instance.profile).to.eq('my-fancy-instance');
      });
  });

  it('UC-SW-16 setup-wizard-identity-inline-validation blocks next on missing instance name', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-wizard-error"]')
      .should('be.visible')
      .and('contain.text', 'Instance name is required.');
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 1 OF 3');
  });

  it('UC-SW-03 setup-wizard-step-controls preserves step form state while navigating previous/next', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Wave3 Instance');
    cy.get('[data-testid="setup-profile-key"]').clear().type('wave3-profile');
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-auth-mode"]').select('clerk');
    cy.get('[data-testid="setup-clerk-publishable-key"]').type('pk_test_wave3');
    cy.get('[data-testid="setup-prev"]').click();

    cy.get('[data-testid="setup-instance-name"]').should(
      'have.value',
      'Wave3 Instance'
    );
    cy.get('[data-testid="setup-profile-key"]').should(
      'have.value',
      'wave3-profile'
    );

    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-auth-mode"]').should('have.value', 'clerk');
    cy.get('[data-testid="setup-clerk-publishable-key"]').should(
      'have.value',
      'pk_test_wave3'
    );
  });

  it('UC-SW-07 setup-wizard-complete-to-launch transitions to config complete with start action', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 3 OF 3');
    cy.get('[data-testid="setup-complete"]').should('be.visible').click();

    cy.contains('Config complete').should('be.visible');
    cy.contains(/Start App|Open Cabinet/).should('be.visible');
    cy.contains(/registration-success|email activation/i).should('not.exist');
  });

  it('UC-SW-04 setup-wizard-completion-state shows runtime and storage details with start action', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.contains('Config complete').should('be.visible');
    cy.contains('Start App').should('be.visible');
    cy.get('[data-testid="setup-complete-config-path"]').should('be.visible');
    cy.get('[data-testid="setup-complete-data-dir"]').should('be.visible');
    cy.get('[data-testid="setup-complete-media-dir"]').should('be.visible');
    cy.get('[data-testid="setup-complete-runtime-url"]').should('contain.text', 'http://');
    cy.get('[data-testid="setup-complete-runtime-port"]').should(($el) => {
      const numeric = Number($el.text().trim());
      expect(Number.isFinite(numeric)).to.eq(true);
      expect(numeric).to.be.greaterThan(0);
    });
  });

  it('UC-SW-08 setup-wizard-config-schema-write persists deterministic cabinet.json payload', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Primary');
    cy.get('[data-testid="setup-profile-key"]').clear().type('primary');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.instance.name).to.eq('Primary');
        expect(payload.instance.profile).to.eq('primary');
        expect(payload.storage.dataDir).to.be.a('string').and.not.empty;
        expect(payload.storage.mediaDir).to.be.a('string').and.not.empty;
        expect(payload.runtime.portMode).to.be.oneOf(['auto', 'fixed']);
        expect(payload.runtime.resolvedUrl).to.match(
          /^http:\/\/(127\.0\.0\.1|0\.0\.0\.0):/
        );
        expect(payload.auth.mode).to.eq('local');
        expect(payload.auth.clerk.enabled).to.eq(false);
        expect(payload.bootstrap.workspace).to.be.a('string').and.not.empty;
        expect(payload.features.chat).to.eq(true);
        expect(payload.features.providers).to.eq(true);
        expect(payload.features.scanner).to.eq(true);
        expect(payload.meta.wizardVersion).to.eq('1');
      });
  });

  it('UC-SW-09 setup-wizard-clerk-required-fields blocks completion when clerk key is missing', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-auth-mode"]').select('clerk');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.get('[data-testid="setup-wizard-error"]')
      .should('be.visible')
      .and('contain.text', 'Clerk publishable key is required');

    cy.request({
      method: 'GET',
      url: '/api/test/runtime/setup-config',
      failOnStatusCode: false,
    }).then((response) => {
      expect(response.status).to.eq(404);
    });

    cy.get('[data-testid="setup-prev"]').click();
    cy.get('[data-testid="setup-clerk-publishable-key"]')
      .clear()
      .type('pk_test_123');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');
  });

  it('UC-SW-05 setup-wizard-not-in-home-shell keeps starter setup controls out of authenticated home', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in');
    cy.get('input[name="email"]').type('e2e-setup-home@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.contains('button', 'Sign in').click();
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^(\/|\/_authenticated\/?)$/
    );
    cy.contains('Home').should('be.visible');

    cy.get('[data-testid="setup-wizard"]').should('not.exist');
    cy.contains('Starter Onboarding').should('not.exist');
    cy.contains('Start Setup').should('not.exist');
    cy.contains('Import Existing Collection').should('not.exist');
    cy.contains('Use Sample Data').should('not.exist');
  });
});
