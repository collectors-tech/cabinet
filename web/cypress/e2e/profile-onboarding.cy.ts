describe("Profile onboarding", () => {
  it("creates and activates first profile from UI", () => {
    cy.visit("/");
    cy.contains("Create First Profile").click();
    cy.contains(/active profile:/i).should("exist");
    cy.contains(/database:/i).should("exist");
    cy.contains(/media:/i).should("exist");
  });
});
