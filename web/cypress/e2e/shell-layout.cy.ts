describe("Workspace shell layout", () => {
  it("shows a three-pane desktop shell with context selection", () => {
    cy.viewport(1440, 900);
    cy.visit("/");

    cy.get("aside.primary-nav").should("exist");
    cy.get('[aria-label="Collection context pane"]').should("exist");
    cy.get('[aria-label="Primary content"]').should("exist");

    cy.contains('[aria-label="Collection context pane"] button', "All Items").click();
    cy.contains('[aria-label="Primary content"]', "Context: All Items").should("exist");
  });

  it("shows context navigation in mobile drawer", () => {
    cy.viewport(390, 844);
    cy.visit("/");

    cy.get('button[aria-label="Open navigation menu"]').click();
    cy.get(".cabinet-drawer").should("exist");
    cy.get('.cabinet-drawer [aria-label="Collection context pane"]').should("exist");
  });
});
