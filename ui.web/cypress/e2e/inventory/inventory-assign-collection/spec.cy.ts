describe("inventory assign to collection", () => {
  const items = [
    {
      id: "item-assign-alpha",
      part_number: "PN-ASSIGN-A",
      title: "Assign Alpha",
      status: "active",
      category: "Cards",
      brand: "Topps",
      priority: "medium",
      description: "Inventory row assignment coverage item",
    },
  ];

  function signIn(path = "/inventory/") {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items },
    }).as("items");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path });
    });
    cy.wait("@items");
  }

  it("assigns an inventory row to a collection and reflects it on Collections", () => {
    signIn();

    cy.get(
      '[data-testid="inventory-item-row-item-assign-alpha"] [data-testid="inventory-row-assign-collection-action"]'
    ).click();
    cy.get('[data-testid="inventory-assign-collection-dialog"]')
      .should("be.visible")
      .and("contain", "Assign Alpha");
    cy.get('[data-testid="inventory-assign-collection-select"]').select("Store 1");
    cy.get('[data-testid="inventory-assign-collection-submit"]').click();
    cy.get('[data-testid="inventory-assign-collection-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-ASSIGN-A");

    cy.reload();
    cy.wait("@items");
    cy.get('[data-testid="inventory-collection-filter-select"]').select("Store 1");
    cy.contains("Assign Alpha").should("be.visible");

    cy.visit("/collections/");
    cy.get('[data-testid="collections-row-store-1"]').click();
    cy.get('[data-testid="collections-member-assign-alpha"]').should("be.visible");
  });
});
