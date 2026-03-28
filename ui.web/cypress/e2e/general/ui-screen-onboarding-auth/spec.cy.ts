describe('UI-SCREEN-ONBOARDING-AUTH', () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
  });

  it('UI-SCREEN-ONBOARDING-AUTH-001 locks workspace until sign-in then unlocks redirect target', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in');

    cy.location('pathname').should('eq', '/sign-in');
    cy.contains('Sign in').should('be.visible');

    cy.get('input[name="email"]').type('e2e-onboarding@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.contains('button', 'Sign in').click();

    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^(\/|\/_authenticated\/?)$/
    );
    cy.contains('Home').should('be.visible');
    cy.contains('Starter Onboarding').should('not.exist');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-003 keeps user on auth screen for invalid credentials input state', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in');

    cy.get('input[name="email"]').type('not-an-email');
    cy.get('input[name="password"]').type('short');
    cy.contains('button', 'Sign in').click();

    cy.location('pathname').should('eq', '/sign-in');
    cy.contains(/invalid email|please enter your email/i).should('be.visible');
    cy.contains('Password must be at least 7 characters long').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-002 resumes persisted onboarding step after reload', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in?redirect=%2F');
    cy.get('input[name="email"]').type('e2e-onboarding-resume@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.contains('button', 'Sign in').click();

    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^(\/|\/_authenticated\/?)$/
    );

    cy.contains('Home').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-009 shows full-screen setup wizard before auth when setup config is missing', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in');
    cy.contains('Setup Wizard').should('be.visible');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.get('[data-testid="setup-start-app"]').click();
    cy.contains('Sign in').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-010 exposes a visible Create account path from sign-in', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in');
    cy.contains('a', 'Create account')
      .should('be.visible')
      .and('have.attr', 'href', '/sign-up')
      .click();
    cy.location('pathname').should('match', /^\/sign-up\/?$/);
  });

  it('UI-SCREEN-ONBOARDING-AUTH-006 renders Google, Apple, and Microsoft provider actions deterministically', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in');
    cy.get('[data-testid="provider-google"]').should('be.visible');
    cy.get('[data-testid="provider-apple"]').should('be.visible');
    cy.get('[data-testid="provider-microsoft"]').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-007 resolves identity mode and provider enablement from runtime config', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.request('POST', '/api/test/auth/provider-options', {
      identity_mode: 'clerk',
      providers: [
        { id: 'google', enabled: true },
        { id: 'apple', enabled: false },
        { id: 'microsoft', enabled: true },
      ],
    })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in');
    cy.get('[data-testid="identity-mode-indicator"]').should('contain.text', 'clerk');
    cy.get('[data-testid="provider-google"]').should('be.visible').and('not.be.disabled');
    cy.get('[data-testid="provider-apple"]').should('be.visible').and('be.disabled');
    cy.get('[data-testid="provider-microsoft"]').should('be.visible').and('not.be.disabled');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-008 signs in with passkey and redirects without password prompt', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in', {
      onBeforeLoad(win) {
        // Minimal passkey-capable environment for deterministic E2E.
        (win as Window & { PublicKeyCredential?: unknown }).PublicKeyCredential =
          function PublicKeyCredential() {
            return undefined;
          };
        Object.defineProperty(win.navigator, 'credentials', {
          configurable: true,
          value: {
            get: () => Promise.resolve({ id: 'e2e-passkey-credential' }),
          },
        });
      },
    });

    cy.get('[data-testid="passkey-signin"]').click();
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^(\/|\/_authenticated\/?)$/
    );
    cy.contains('Home').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-008 shows deterministic fallback guidance when passkey is unavailable', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-in', {
      onBeforeLoad(win) {
        (win as Window & { PublicKeyCredential?: unknown }).PublicKeyCredential =
          undefined;
        Object.defineProperty(win.navigator, 'credentials', {
          configurable: true,
          value: undefined,
        });
      },
    });
    cy.get('[data-testid="passkey-signin"]').click();
    cy.get('[data-testid="passkey-error"]')
      .should('be.visible')
      .and(
        'contain.text',
        'Passkey sign-in is unavailable on this device. Use password or provider sign-in.'
      );
    cy.contains('button', 'Sign in').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-011 completes sign-up and redirects to authenticated shell', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/sign-up');
    cy.get('input[name="email"]').type('e2e-signup@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.get('input[name="confirmPassword"]').type('password123');

    cy.contains('button', 'Create Account').click();

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/dashboard\/?$/);
    cy.contains('Home').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-012 completes forgot-password submit and routes to OTP recovery', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/forgot-password');
    cy.get('input[name="email"]').type('valid.user@example.com');
    cy.contains('button', 'Continue').click();

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/otp\/?$/);
    cy.contains(/please enter your email/i).should('not.exist');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-013 renders Privacy Policy content on the public privacy route', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/privacy');
    cy.contains('h1', 'Privacy Policy').should('be.visible');
    cy.contains('What Cabinet stores').should('be.visible');
    cy.contains('Optional integrations').should('be.visible');
    cy.contains(/^404$/).should('not.exist');
    cy.contains('Go Back').should('not.exist');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-013 renders Terms of Service content on the public terms route', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/terms');
    cy.contains('h1', 'Terms of Service').should('be.visible');
    cy.contains('Acceptable use').should('be.visible');
    cy.contains('Account and data responsibility').should('be.visible');
    cy.contains(/^404$/).should('not.exist');
    cy.contains('Go Back').should('not.exist');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-014 keeps resend in OTP context with deterministic feedback', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.visit('/otp');
    cy.get('[data-testid="otp-resend"]').click();
    cy.location('pathname').should('match', /^\/otp\/?$/);
    cy.get('[data-testid="otp-resend-feedback"]')
      .should('contain', 'A new verification code was sent.')
      .and('contain', 'Stay on this screen and enter the latest code.');
    cy.contains('Sign in').should('not.exist');
  });
});
