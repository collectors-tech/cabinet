describe('UI-SCREEN-ONBOARDING-AUTH', () => {
  it('UI-SCREEN-ONBOARDING-AUTH-001 locks workspace until sign-in then unlocks redirect target', () => {
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
  });

  it('UI-SCREEN-ONBOARDING-AUTH-003 keeps user on auth screen for invalid credentials input state', () => {
    cy.visit('/sign-in');

    cy.get('input[name="email"]').type('not-an-email');
    cy.get('input[name="password"]').type('short');
    cy.contains('button', 'Sign in').click();

    cy.location('pathname').should('eq', '/sign-in');
    cy.contains(/invalid email|please enter your email/i).should('be.visible');
    cy.contains('Password must be at least 7 characters long').should('be.visible');
  });

  it('UI-SCREEN-ONBOARDING-AUTH-002 resumes persisted onboarding step after reload', () => {
    cy.visit('/sign-in?redirect=%2F');
    cy.get('input[name="email"]').type('e2e-onboarding-resume@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.contains('button', 'Sign in').click();

    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^(\/|\/_authenticated\/?)$/
    );

    cy.get('[data-testid="onboarding-step-label"]').should('contain', 'Welcome');
    cy.get('[data-testid="onboarding-next-step"]').click();
    cy.get('[data-testid="onboarding-step-label"]').should('contain', 'Identity');

    cy.reload();
    cy.get('[data-testid="onboarding-step-label"]').should('contain', 'Identity');
  });
});
