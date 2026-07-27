describe('SETUP-WIZ', () => {
  function enterSetupFormMode() {
    cy.get('[data-testid="setup-start"]').click();
    cy.get('[data-testid="setup-instance-name"]').should('be.visible');
  }

  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
    cy.e2eSetSetupState('missing');
    cy.intercept('GET', '/api/runtime/setup-status').as('setupStatus');
    cy.visit('/sign-in');
    cy.wait('@setupStatus');
    cy.get('[data-testid="setup-wizard"]').should('be.visible');
  });

  it('UC-SW-35 setup-helper-bypass-seeds-config before route assertions', () => {
    cy.e2eSetSetupState('present');
    cy.visit('/sign-in');
    cy.get('[data-testid="setup-wizard"]').should('not.exist');
    cy.contains('Sign in').should('be.visible');
  });

  it('UC-SW-36 setup-helper-completion-path clears setup gate deterministically', () => {
    cy.e2eSetSetupState('missing');
    cy.e2eCompleteSetupHelper({
      instance_name: 'E2E Setup Completion',
      profile_key: 'e2e-setup-completion',
    });
    cy.visit('/sign-in');
    cy.get('[data-testid="setup-wizard"]').should('not.exist');
    cy.contains('Sign in').should('be.visible');
  });

  it('UC-SW-06 setup-wizard-progress-template shows step header, percentage, and footer actions', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 1 OF 6');
    cy.get('[data-testid="setup-step-percent"]').should('contain.text', '17%');
    cy.get('[data-testid="setup-progress-bar"]')
      .should('have.attr', 'aria-valuenow')
      .and('eq', '17');
    cy.get('[data-testid="setup-prev"]').should('be.disabled');
    cy.get('[data-testid="setup-next"]').should('be.visible').and('not.be.disabled');
    cy.get('[data-testid="setup-complete"]').should('not.exist');

    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 2 OF 6');
    cy.get('[data-testid="setup-step-percent"]').should('contain.text', '33%');
    cy.get('[data-testid="setup-prev"]').should('not.be.disabled');
  });

  it('UC-SW-12 setup-wizard-welcome-actions renders start/import actions before form fields', () => {
    cy.get('[data-testid="setup-start"]').should('be.visible');
    cy.get('[data-testid="setup-use-defaults"]').should('be.visible');
    cy.get('[data-testid="setup-import-toggle"]').should('be.visible');
    cy.get('[data-testid="setup-instance-name"]').should('not.exist');
    cy.get('[data-testid="setup-profile-key"]').should('not.exist');

    cy.get('[data-testid="setup-start"]').click();
    cy.get('[data-testid="setup-instance-name"]').should('be.visible');
    cy.get('[data-testid="setup-profile-key"]').should('be.visible');
  });

  it('UC-SW-37 setup-wizard-use-defaults writes deterministic config and shows defaults-applied completion feedback', () => {
    cy.get('[data-testid="setup-use-defaults"]').click();
    cy.get('[data-testid="setup-wizard-complete-state"]').should('be.visible');
    cy.get('[data-testid="setup-complete-feedback"]')
      .should('be.visible')
      .and('contain.text', 'Defaults applied.');
    cy.get('[data-testid="setup-complete-instance-name"]').should(
      'contain.text',
      'Cabinet Local'
    );
    cy.get('[data-testid="setup-complete-runtime-port"]').should(($el) => {
      const numeric = Number($el.text().trim());
      expect(Number.isFinite(numeric)).to.eq(true);
      expect(numeric).to.be.greaterThan(0);
    });

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.instance.name).to.eq('Cabinet Local');
        expect(payload.instance.profile).to.eq('default');
        expect(payload.auth.mode).to.eq('local');
      });
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
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 6 OF 6');
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
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 1 OF 6');
  });

  it('UC-SW-17 setup-wizard-storage-defaults shows exe-local mode and default data path', () => {
    cy.request('GET', '/api/runtime/setup-status').then((statusResponse) => {
      enterSetupFormMode();
      cy.get('[data-testid="setup-next"]').click();

      cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 2 OF 6');
      cy.get('[data-testid="setup-storage-mode"]').should('have.value', 'exe_local');
      cy.get('[data-testid="setup-storage-mode"] option:selected').should(
        'have.text',
        'local'
      );
      cy.get('[data-testid="setup-storage-data-dir-preview"]')
        .should('be.visible')
        .and('contain.text', statusResponse.body.default_storage_data_dir);
      cy.get('[data-testid="setup-storage-portable-mode"]').should('be.visible');
      cy.get('[data-testid="setup-storage-portable-mode"]').should('not.be.checked');
    });
  });

  it('UC-SW-18 setup-wizard-storage-custom-path-validation blocks blank custom path', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-storage-mode"]').select('custom');
    cy.get('[data-testid="setup-storage-custom-data-dir"]').clear();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-wizard-error"]')
      .should('be.visible')
      .and('contain.text', 'Custom storage path is required.');
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 2 OF 6');
  });

  it('UC-SW-19 setup-wizard-storage-selection persists data and media dirs', () => {
    const customDir = 'C:\\projects\\collectors-tech\\cabinet\\tmp\\cabinet-e2e-storage';

    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Storage Persistence');
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-storage-mode"]').select('custom');
    cy.get('[data-testid="setup-storage-custom-data-dir"]').clear().type(customDir);
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 6 OF 6');
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.storage.dataDir).to.eq(customDir);
        expect(payload.storage.mediaDir).to.eq(`${customDir}\\media`);
      });
  });

  it('UC-SW-20 setup-wizard-runtime-defaults shows auto mode with resolved URL preview', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 3 OF 6');
    cy.get('[data-testid="setup-runtime-port-mode"]').should('have.value', 'auto');
    cy.get('[data-testid="setup-runtime-url-preview"]')
      .should('be.visible')
      .and('contain.text', 'http://');
  });

  it('UC-SW-21 setup-wizard-runtime-fixed-port-validation blocks invalid fixed port', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-runtime-port-mode"]').select('fixed');
    cy.get('[data-testid="setup-runtime-fixed-port"]').type('{selectall}0');
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-wizard-error"]')
      .should('be.visible')
      .and('contain.text', 'Fixed port value is required');
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 3 OF 6');
  });

  it('UC-SW-22 setup-wizard-runtime-selection persists fixed port and resolved URL', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Runtime Persist');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-runtime-port-mode"]').select('fixed');
    cy.get('[data-testid="setup-runtime-fixed-port"]').type('{selectall}18999');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.runtime.portMode).to.eq('fixed');
        expect(payload.runtime.port).to.eq(18999);
        expect(payload.runtime.resolvedUrl).to.match(/:18999$/);
      });
  });

  it('UC-SW-23 setup-wizard-auth-mode-switch exposes only local and ZITADEL readiness state', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-auth-mode"]').should('have.value', 'local');
    cy.get('[data-testid="setup-auth-mode"] option[value="clerk"]').should('not.exist');
    cy.get('[data-testid="setup-clerk-publishable-key"]').should('not.exist');
    cy.get('[data-testid="setup-auth-readiness"]').should('contain.text', 'Ready: local-device mode');

    cy.get('[data-testid="setup-auth-mode"]').select('zitadel');
    cy.get('[data-testid="setup-clerk-publishable-key"]').should('not.exist');
    cy.get('[data-testid="setup-clerk-built-in-key"]').should('not.exist');
    cy.get('[data-testid="setup-auth-readiness"]')
      .should('contain.text', 'ZITADEL application')
      .and('contain.text', 'callback');

    cy.get('[data-testid="setup-auth-mode"]').select('local');
    cy.get('[data-testid="setup-clerk-publishable-key"]').should('not.exist');
    cy.get('[data-testid="setup-auth-readiness"]').should('contain.text', 'Ready: local-device mode');
  });

  it('UC-SW-24 setup-wizard-zitadel-readiness allows next without Clerk key entry', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-auth-mode"]').select('zitadel');
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-wizard-error"]').should('not.exist');
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 5 OF 6');
  });

  it('SETUP-WIZ-015 + #1968 keeps retired Clerk key out of setup payload', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('ZITADEL Setup');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-auth-mode"]').select('zitadel');
    cy.get('[data-testid="setup-clerk-publishable-key"]').should('not.exist');
    cy.get('[data-testid="setup-clerk-built-in-key"]').should('not.exist');
    cy.get('[data-testid="setup-auth-readiness"]')
      .should('contain.text', 'ZITADEL application')
      .and('contain.text', 'required role grants');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.auth.mode).to.eq('zitadel');
        expect(payload.auth.clerk.enabled).to.eq(false);
        expect(payload.auth.clerk.publishableKey).to.eq('');
      });
  });

  it('UC-SW-25 setup-wizard-auth-readiness-configured persists ZITADEL auth config', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('ZITADEL Persist');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-auth-mode"]').select('zitadel');
    cy.get('[data-testid="setup-clerk-publishable-key"]').should('not.exist');
    cy.get('[data-testid="setup-clerk-built-in-key"]').should('not.exist');
    cy.get('[data-testid="setup-auth-readiness"]')
      .should('contain.text', 'ZITADEL application')
      .and('contain.text', 'required role grants');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.auth.mode).to.eq('zitadel');
        expect(payload.auth.clerk.enabled).to.eq(false);
        expect(payload.auth.clerk.publishableKey).to.eq('');
      });
  });

  it('UC-SW-26 setup-wizard-integrations-defaults shows enabled toggles and guidance', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 5 OF 6');
    cy.get('[data-testid="setup-integrations-guidance"]')
      .should('be.visible')
      .and('contain.text', 'edit integrations any time later in Settings');
    cy.get('[data-testid="setup-feature-scanner"]').should('be.checked');
    cy.get('[data-testid="setup-feature-chat"]').should('be.checked');
    cy.get('[data-testid="setup-feature-providers"]').should('be.checked');
  });

  it('UC-SW-27 setup-wizard-integrations-persistence writes feature toggles', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Feature Persist');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-feature-scanner"]').uncheck({ force: true });
    cy.get('[data-testid="setup-feature-chat"]').uncheck({ force: true });
    cy.get('[data-testid="setup-feature-providers"]').check({ force: true });
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.features.scanner).to.eq(false);
        expect(payload.features.chat).to.eq(false);
        expect(payload.features.providers).to.eq(true);
      });
  });

  it('UC-SW-28 setup-wizard-integrations-optional-step-allows-next', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 6 OF 6');
    cy.get('[data-testid="setup-wizard-error"]').should('not.exist');
  });

  it('UC-SW-29 setup-wizard-review-summary shows resolved full summary and create label', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Review Summary');
    cy.get('[data-testid="setup-profile-key"]').clear().type('review-summary');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-runtime-port-mode"]').select('fixed');
    cy.get('[data-testid="setup-runtime-fixed-port"]').type('{selectall}18888');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-auth-mode"]').select('zitadel');
    cy.get('[data-testid="setup-clerk-built-in-key"]').should('not.exist');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-feature-chat"]').uncheck({ force: true });
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 6 OF 6');
    cy.contains('strong', 'Instance:').should('be.visible');
    cy.contains('strong', 'Storage Mode:').should('be.visible');
    cy.contains('strong', 'Runtime URL:').should('be.visible');
    cy.contains('strong', 'Auth Mode:').should('be.visible');
    cy.contains('strong', 'Features:').should('be.visible');
    cy.get('[data-testid="setup-complete"]').should('contain.text', 'Create Config & Launch');
  });

  it('UC-SW-30 setup-wizard-review-create-action writes config and completion metadata', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Review Persist');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.contains('Config complete').should('be.visible');
    cy.get('[data-testid="setup-complete-config-path"]').should('be.visible');
    cy.get('[data-testid="setup-complete-runtime-url"]').should('contain.text', 'http://');
    cy.get('[data-testid="setup-complete-runtime-port"]').should('be.visible');
    cy.get('[data-testid="setup-complete-local-credentials"]').should('be.visible');
    cy.get('[data-testid="setup-complete-local-credentials"]').should(
      'contain.text',
      'Local device mode'
    );
  });

  it('UC-SW-31 setup-wizard-review-create-action shows in-flight disabled state', () => {
    cy.intercept('POST', '/api/runtime/setup-complete', {
      delay: 1200,
      statusCode: 200,
      body: {
        ok: true,
        config_path: 'D:/cabinet.json',
        data_dir: 'D:/data',
        media_dir: 'D:/data/media',
        runtime_url: 'http://127.0.0.1:17880',
        runtime_port: 17880,
      },
    }).as('setupCompleteSlow');

    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.get('[data-testid="setup-complete"]')
      .should('be.disabled')
      .and('contain.text', 'Creating Config...');
    cy.wait('@setupCompleteSlow');
    cy.contains('Config complete').should('be.visible');
  });

  it('UC-SW-32 setup-wizard-completion-summary shows launch confirmation actions', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Launch Summary');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.contains('Config complete').should('be.visible');
    cy.get('[data-testid="setup-complete-runtime-url"]').should('contain.text', 'http://');
    cy.get('[data-testid="setup-complete-instance-name"]').should('contain.text', 'Launch Summary');
    cy.get('[data-testid="setup-complete-data-dir"]').should('be.visible');
    cy.get('[data-testid="setup-open-cabinet"]').should('be.visible');
    cy.get('[data-testid="setup-open-config-folder"]').should('be.visible');
    cy.get('[data-testid="setup-finish"]').should('be.visible');
  });

  it('UC-SW-33 setup-wizard-open-cabinet exits completion state', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.contains('Config complete').should('be.visible');
    cy.get('[data-testid="setup-open-cabinet"]').click();
    cy.get('[data-testid="setup-wizard"]').should('not.exist');
    cy.contains('Sign in').should('be.visible');
  });

  it('UC-SW-34 setup-wizard-open-config-folder shows feedback', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.contains('Config complete').should('be.visible');
    cy.get('[data-testid="setup-open-config-folder"]').click();
    cy.get('[data-testid="setup-complete-feedback"]')
      .should('be.visible')
      .and('contain.text', 'Config folder');
  });

  it('UC-SW-38 setup-wizard-local-completion-shows-working-login-credentials', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Local Credentials');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-auth-mode"]').should('have.value', 'local');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.get('[data-testid="setup-complete-local-credentials"]').should('be.visible');
    cy.get('[data-testid="setup-complete-local-credentials"]').should(
      'contain.text',
      'Local device mode'
    );

    cy.get('[data-testid="setup-open-cabinet"]').click();
    cy.location('pathname', { timeout: 15000 }).should('eq', '/sign-in');
    cy.get('[data-testid="setup-wizard"]').should('not.exist');
  });

  it('UC-SW-03 setup-wizard-step-controls preserves step form state while navigating previous/next', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-instance-name"]').clear().type('Wave3 Instance');
    cy.get('[data-testid="setup-profile-key"]').clear().type('wave3-profile');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();

    cy.get('[data-testid="setup-auth-mode"]').select('zitadel');
    cy.get('[data-testid="setup-clerk-built-in-key"]').should('not.exist');
    cy.get('[data-testid="setup-prev"]').click();
    cy.get('[data-testid="setup-prev"]').click();
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
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-auth-mode"]').should('have.value', 'zitadel');
    cy.get('[data-testid="setup-clerk-publishable-key"]').should('not.exist');
    cy.get('[data-testid="setup-clerk-built-in-key"]').should('not.exist');
  });

  it('UC-SW-07 setup-wizard-complete-to-launch transitions to config complete with start action', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 6 OF 6');
    cy.get('[data-testid="setup-complete"]').should('be.visible').click();

    cy.contains('Config complete').should('be.visible');
    cy.contains(/Start App|Open Cabinet/).should('be.visible');
    cy.contains(/registration-success|email activation/i).should('not.exist');
  });

  it('UC-SW-04 setup-wizard-completion-state shows runtime and storage details with start action', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.contains('Config complete').should('be.visible');
    cy.contains('Open Cabinet').should('be.visible');
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
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.instance.name).to.eq('Primary');
        expect(payload.instance.profile).to.eq('primary');
        expect(String(payload.storage.dataDir).trim()).not.to.equal('');
        expect(String(payload.storage.mediaDir).trim()).not.to.equal('');
        expect(payload.runtime.portMode).to.be.oneOf(['auto', 'fixed']);
        expect(payload.runtime.resolvedUrl).to.match(
          /^http:\/\/(127\.0\.0\.1|0\.0\.0\.0):/
        );
        expect(payload.auth.mode).to.eq('local');
        expect(payload.auth.clerk.enabled).to.eq(false);
        expect(String(payload.bootstrap.workspace).trim()).not.to.equal('');
        expect(payload.features.chat).to.eq(true);
        expect(payload.features.providers).to.eq(true);
        expect(payload.features.scanner).to.eq(true);
        expect(payload.meta.wizardVersion).to.eq('1');
      });
  });

  it('UC-SW-09 setup-wizard-zitadel-completes without retired Clerk key', () => {
    enterSetupFormMode();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-auth-mode"]').select('zitadel');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.auth.mode).to.eq('zitadel');
        expect(payload.auth.clerk.enabled).to.eq(false);
        expect(payload.auth.clerk.publishableKey).to.eq('');
      });
  });

  it('UC-SW-05 setup-wizard-not-in-home-shell keeps starter setup controls out of authenticated home', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in');
    cy.location('pathname', { timeout: 15000 }).should('eq', '/sign-in');

    cy.get('[data-testid="setup-wizard"]').should('not.exist');
    cy.contains('Starter Onboarding').should('not.exist');
    cy.contains('Start Setup').should('not.exist');
    cy.contains('Import Existing Collection').should('not.exist');
    cy.contains('Use Sample Data').should('not.exist');
  });
});

