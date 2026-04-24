describe("inventory-compact-collection-filter", () => {
  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
  }

  it("uses compact collection filters instead of the folder tree panel", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-filter-1",
            part_number: "PN-FILTER-1",
            title: "Store One Filtered Item",
            status: "active",
            category: "Cars",
          },
          {
            id: "item-filter-2",
            part_number: "PN-FILTER-2",
            title: "Watch List Item",
            status: "active",
            category: "Cars",
          },
        ],
      },
    }).as("itemsCollectionFilter");

    signIn();
    cy.wait("@itemsCollectionFilter");
    cy.window().then((win) => {
      win.localStorage.setItem(
        "cabinet.inventory.item-folder-assignments.v1",
        JSON.stringify({
          "item-filter-1": "Store 1",
          "item-filter-2": "Watch List",
        })
      );
    });
    cy.reload();
    cy.wait("@itemsCollectionFilter");

    cy.get('[data-testid="inventory-folder-tree"]').should("not.exist");
    cy.get('[data-testid="inventory-collection-filter"]').should("be.visible");
    cy.get('[data-testid="inventory-collection-filter-selected"]').should(
      "contain",
      "All Items"
    );
    cy.contains("Store One Filtered Item").should("be.visible");
    cy.contains("Watch List Item").should("be.visible");

    cy.get('[data-testid="inventory-collection-filter-select"]').select("Store 1");
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 1");
    cy.get('[data-testid="inventory-collection-filter-selected"]').should(
      "contain",
      "Store 1"
    );
    cy.contains("Store One Filtered Item").should("be.visible");
    cy.contains("Watch List Item").should("not.exist");

    cy.get('[data-testid="inventory-collection-add-root"]').click();
    cy.get('[data-testid="folder-tree-name-input"]').type("Empty Filter Bucket");
    cy.get('[data-testid="folder-tree-create-submit"]').click();
    cy.get('[data-testid="collection-active-context"]').should(
      "contain",
      "Empty Filter Bucket"
    );
    cy.contains("Store One Filtered Item").should("not.exist");
    cy.contains("Watch List Item").should("not.exist");
    cy.contains("No results.").should("be.visible");
  });
});
