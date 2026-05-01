describe("inventory collection browser picker actions", () => {
  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-context-alpha",
            part_number: "PN-CONTEXT-A",
            title: "Context Alpha",
            status: "active",
            category: "Cards",
          },
        ],
      },
    }).as("items");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
    cy.wait("@items");
  }

  it("keeps Browse as a picker without management row actions", () => {
    signIn();

    cy.get('[data-testid="inventory-collection-browser-trigger"]').click();
    cy.get('[data-testid="inventory-folder-browser-menu"]')
      .should("be.visible")
      .within(() => {
        cy.get('[data-testid="folder-tree-row-actions-store-1"]').should(
          "not.exist"
        );
        cy.get('[data-testid="folder-tree-add-child-store-1"]').should(
          "not.exist"
        );
        cy.get('[data-testid="folder-tree-row-action-properties-store-1"]').should(
          "not.exist"
        );
      });
    cy.get('[data-testid="inventory-folder-browser-menu"]')
      .find('[data-testid="folder-tree-item-store-1"]')
      .click();
    cy.get('[data-testid="inventory-collection-filter-selected"]').should(
      "contain",
      "Store 1"
    );
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 1");
    cy.contains("No results.").should("be.visible");
    cy.contains("Context Alpha").should("not.exist");

    cy.get('[data-testid="inventory-folder-tree"]')
      .find('[data-testid="folder-tree-row-actions-store-1"]')
      .click();
    cy.get('[data-testid="folder-tree-row-action-properties-store-1"]').click();
    cy.get('[data-testid="folder-properties-name-input"]').should("have.value", "Store 1");
  });
});
