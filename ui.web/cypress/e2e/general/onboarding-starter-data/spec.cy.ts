describe('ONBOARDING-STARTER-DATA', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2F');
    cy.get('input[name="email"]').clear().type('e2e-onboarding-starter@example.com');
    cy.get('input[name="password"]').clear().type('password123');
    cy.contains('button', 'Sign in').click();
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^(\/|\/_authenticated\/?)$/
    );
  }

  it('ONBOARDING-STARTER-DATA-001 shows guided starter onboarding controls without advanced-form lock-in', () => {
    signIn();

    cy.contains('Starter Onboarding').should('be.visible');
    cy.contains('button', 'Start Setup').should('be.visible');
    cy.contains('Import Existing Collection').should('be.visible');
    cy.contains('button', 'Use Sample Data').should('be.visible');
    cy.contains(/Brand|Part Number|Acquisition price/i).should('not.exist');
  });

  it('ONBOARDING-STARTER-DATA-002 seeds sample data and displays deterministic summary counts', () => {
    cy.intercept('POST', '/api/onboarding/sample-data', {
      statusCode: 200,
      body: {
        already_seeded_for_profile: false,
        folders_created: 3,
        items_created: 6,
        media_created: 0,
      },
    }).as('seedSample');

    signIn();
    cy.contains('button', 'Use Sample Data').click();
    cy.wait('@seedSample');

    cy.get('[data-testid="onboarding-seed-summary"]')
      .should('be.visible')
      .and('contain', 'Folders: 3')
      .and('contain', 'Items: 6')
      .and('contain', 'Media: 0');
  });

  it('ONBOARDING-STARTER-DATA-003 routes to import flow without auto-seeding sample data', () => {
    cy.intercept('POST', '/api/onboarding/sample-data').as('seedSample');

    signIn();
    cy.contains('Import Existing Collection').click();

    cy.location('pathname', { timeout: 10000 }).should(
      'match',
      /^\/settings\/storage\/?$/
    );
    cy.get('@seedSample.all').should('have.length', 0);
  });
});
