describe("inventory-shell", () => {
  function openInventory() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.e2eEnsureSignedOut();
      cy.stubLocalServerSession(profile_id);
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/inventory/",
      });
    });
  }

  it("UI-SCREEN-INVENTORY-SHELL-001 renders resolved inventory intro copy", () => {
    const inventoryIntro =
      "Browse, organize, and update the items you already own.";

    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsIntro");

    openInventory();
    cy.wait("@itemsIntro");

    cy.get('[data-testid="inventory-header-title"]')
      .should("be.visible")
      .and("have.attr", "title", inventoryIntro)
      .and("have.attr", "aria-label", `Inventory - ${inventoryIntro}`);
    cy.contains(inventoryIntro).should("not.exist");
    cy.contains("inventory.description").should("not.exist");
  });
});
