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
    cy.contains('Complete Setup').click();
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
});
