import "./commands";

before(() => {
  Cypress.on("uncaught:exception", () => false);

  const relative = Cypress.spec.relative || "";
  const setupFlowSpec =
    relative.includes("general/setup-wizard-first-run/spec.cy.ts") ||
    relative.includes("general/ui-screen-onboarding-auth/spec.cy.ts");

  if (!setupFlowSpec) {
    cy.e2eSetSetupState("present");
  }
});
