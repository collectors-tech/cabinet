describe("inventory-compact-collection-filter", () => {
  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
  }

  it("keeps folder context controlled by the tree instead of duplicate toolbar filters", () => {
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

    cy.get('[data-testid="inventory-folder-tree"]').should("be.visible");
    cy.get('[data-testid="folder-tree-item-store-1"]').should("be.visible");
    cy.get('[data-testid="inventory-collection-filter"]').should("not.exist");
    cy.get('[data-testid="inventory-collection-filter-select"]').should("not.exist");
    cy.get('[data-testid="inventory-table-toolbar"]').within(() => {
      cy.contains("button", "Condition").should("not.exist");
      cy.contains("button", "Category").should("not.exist");
    });
    cy.contains("Store One Filtered Item").should("be.visible");
    cy.contains("Watch List Item").should("be.visible");

    cy.get('[data-testid="folder-tree-item-store-1"]').click();
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 1");
    cy.contains("Store One Filtered Item").should("be.visible");
    cy.contains("Watch List Item").should("not.exist");

    cy.get('[data-testid="folder-tree-add-root"]').click();
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

  it("filters by full folder-tree folders instead of falling back to All Items", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-store-3-only",
            part_number: "PN-STORE-3",
            title: "Store Three Exclusive Item",
            status: "active",
            category: "Cars",
          },
          {
            id: "item-warehouse-2-only",
            part_number: "PN-WAREHOUSE-2",
            title: "Warehouse Two Exclusive Item",
            status: "active",
            category: "Cards",
          },
        ],
      },
    }).as("itemsFolderTreeFilter");

    signIn();
    cy.wait("@itemsFolderTreeFilter");
    cy.window().then((win) => {
      win.localStorage.setItem(
        "cabinet.inventory.item-folder-assignments.v1",
        JSON.stringify({
          "item-store-3-only": "Store 3",
          "item-warehouse-2-only": "Warehouse 2",
        })
      );
    });
    cy.reload();
    cy.wait("@itemsFolderTreeFilter");

    cy.get('[data-testid="folder-tree-item-store-3"]').click();
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 3");
    cy.contains("Store Three Exclusive Item").should("be.visible");
    cy.contains("Warehouse Two Exclusive Item").should("not.exist");

    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click();
    cy.get('[data-testid="folder-tree-item-warehouse-2"]').click();
    cy.get('[data-testid="collection-active-context"]').should("contain", "Warehouse 2");
    cy.contains("Warehouse Two Exclusive Item").should("be.visible");
    cy.contains("Store Three Exclusive Item").should("not.exist");
  });

  it("uses the folder tree as the only inventory context filter and keeps tree counts aligned with rows", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-watch-count-1",
            part_number: "PN-WATCH-1",
            title: "Watch Count One",
            status: "active",
            category: "Cars",
          },
          {
            id: "item-watch-count-2",
            part_number: "PN-WATCH-2",
            title: "Watch Count Two",
            status: "active",
            category: "Cars",
          },
          {
            id: "item-store-count-1",
            part_number: "PN-STORE-COUNT-1",
            title: "Store Count One",
            status: "active",
            category: "Cards",
          },
        ],
      },
    }).as("itemsFolderCountFilter");

    signIn();
    cy.wait("@itemsFolderCountFilter");
    cy.window().then((win) => {
      win.localStorage.setItem(
        "cabinet.inventory.item-folder-assignments.v1",
        JSON.stringify({
          "item-watch-count-1": "Watch List",
          "item-watch-count-2": "Watch List",
          "item-store-count-1": "Store 1",
        })
      );
    });
    cy.reload();
    cy.wait("@itemsFolderCountFilter");

    cy.get('[data-testid="inventory-collection-filter-select"]').should("not.exist");
    cy.get('[data-testid="inventory-table-toolbar"]').within(() => {
      cy.contains("button", "Condition").should("not.exist");
      cy.contains("button", "Category").should("not.exist");
    });

    cy.get('[data-testid="folder-tree-item-watch-list"]').click();
    cy.get('[data-testid="folder-tree-count-watch-list"]').should("have.text", "2");
    cy.get('[data-testid="collection-active-context"]').should("contain", "Watch List");
    cy.get('[data-testid^="inventory-item-row-"]').should("have.length", 2);
    cy.contains("Watch Count One").should("be.visible");
    cy.contains("Watch Count Two").should("be.visible");
    cy.contains("Store Count One").should("not.exist");
    cy.get('[data-testid="inventory-selected-folder-empty"]').should("not.exist");
  });
});
