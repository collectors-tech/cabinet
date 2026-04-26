describe("inventory item editor modal", () => {
  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
  }

  it("creates, cancels, edits, saves, and navigates inventory records in a modal", () => {
    const items = [
      {
        id: "item-alpha",
        part_number: "PN-ALPHA",
        title: "Alpha Item",
        status: "active",
        category: "Cars",
        brand: "AFX",
        priority: "medium",
        description: "Alpha description",
      },
      {
        id: "item-bravo",
        part_number: "PN-BRAVO",
        title: "Bravo Item",
        status: "active",
        category: "Trains",
        brand: "Tyco",
        priority: "medium",
        description: "Bravo description",
      },
    ];

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsList");

    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body).to.include({
        part_number: "PN-CREATED",
        title: "Created Modal Item",
      });
      const created = {
        id: "item-created",
        part_number: "PN-CREATED",
        title: "Created Modal Item",
        status: "active",
        category: "Cars",
        brand: "Aurora",
        priority: "medium",
        description: "Created from modal",
      };
      items.unshift(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createItem");

    cy.intercept("PUT", "/api/items/item-bravo", (req) => {
      expect(req.body).to.include({
        title: "Bravo Item Updated",
        brand: "Updated Brand",
      });
      const index = items.findIndex((item) => item.id === "item-bravo");
      items[index] = {
        ...items[index],
        title: "Bravo Item Updated",
        brand: "Updated Brand",
      };
      req.reply({ statusCode: 200, body: items[index] });
    }).as("updateBravo");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-item-editor"]').should("not.exist");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and("contain", "Create Item");
    cy.get('[data-testid="inventory-item-editor-cancel"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.contains("Created Modal Item").should("not.exist");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-part-number"]').clear().type("PN-CREATED");
    cy.get('[data-testid="inventory-item-title"]').clear().type("Created Modal Item");
    cy.get('[data-testid="inventory-item-brand"]').clear().type("Aurora");
    cy.get('[data-testid="inventory-item-category"]').clear().type("Cars");
    cy.get('[data-testid="inventory-item-description"]').clear().type("Created from modal");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@createItem");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-CREATED");
    cy.contains("Created Modal Item").should("be.visible");

    cy.get('[data-testid="inventory-item-row-item-alpha"]').dblclick();
    cy.get('[data-testid="inventory-item-editor-panel"]')
      .should("be.visible")
      .and("contain", "Edit Item");
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Alpha Item");
    cy.get('[data-testid="inventory-item-editor-cancel"]').click();
    cy.get('[data-testid="inventory-item-editor-panel"]').should("not.exist");

    cy.get('[data-testid="inventory-item-row-item-alpha"] [data-testid="task-row-actions-trigger"]').click();
    cy.contains('[role="menuitem"]', "Edit").click();
    cy.get('[data-testid="inventory-item-editor-panel"]')
      .should("be.visible")
      .and("contain", "Edit Item");
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Alpha Item");

    cy.get('[data-testid="inventory-item-editor-next"]').click();
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Bravo Item");
    cy.get('[data-testid="inventory-item-editor-previous"]').click();
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Alpha Item");
    cy.get('[data-testid="inventory-item-editor-next"]').click();
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Bravo Item");
    cy.get('[data-testid="inventory-item-brand"]').clear().type("Updated Brand");
    cy.get('[data-testid="inventory-item-title"]').clear().type("Bravo Item Updated");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@updateBravo");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-panel"]').should("not.exist");
    cy.contains("Bravo Item Updated").should("be.visible");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-BRAVO");
  });
});
