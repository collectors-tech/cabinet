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

    cy.get('[data-testid="inventory-item-row-item-alpha"]').click();
    cy.wait(250);
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-ALPHA");
    cy.get('[data-testid="row-details-modal"]').should("not.exist");
    cy.get('[data-testid="row-edit-modal"]').should("not.exist");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.get('[data-testid="inventory-item-editor-panel"]').should("not.exist");

    cy.get('[data-testid="inventory-item-row-item-alpha"]').dblclick();
    cy.get('[data-testid="inventory-item-editor-panel"]')
      .should("be.visible")
      .and("contain", "Edit Item");
    cy.get('[data-testid="inventory-item-editor-panel"]').then(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.left, "editor panel starts on right half").to.be.greaterThan(
        Cypress.config("viewportWidth") / 2
      );
      expect(rect.right, "editor panel is anchored to right edge").to.be.closeTo(
        Cypress.config("viewportWidth"),
        24
      );
    });
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

  it("uses one create modal for manual, paste, photo, and barcode header actions", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsList");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and("contain", "Create Item")
      .and("not.contain", "Create Item From");
    cy.get('[data-testid="inventory-create-paste-input"]').should("be.visible");
    cy.get('[data-testid="inventory-create-take-image"]')
      .scrollIntoView()
      .should("be.visible");
    cy.get('[data-testid="inventory-create-barcode-input"]').should("be.visible");
    cy.get('[data-testid="inventory-item-editor-cancel"]').click();

    cy.get('[data-testid="inventory-barcodes-action"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and("contain", "Create Item")
      .and("not.contain", "Create Item From Barcode");
    cy.get('[data-testid="inventory-create-paste-input"]').should("be.visible");
    cy.get('[data-testid="inventory-create-take-image"]')
      .scrollIntoView()
      .should("be.visible");
    cy.get('[data-testid="inventory-create-barcode-input"]')
      .should("be.visible")
      .and("be.focused");
    cy.get('[data-testid="inventory-item-editor-cancel"]').click();

    cy.get('[data-testid="inventory-photos-action"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and("contain", "Create Item")
      .and("not.contain", "Create Item From Photo");
    cy.get('[data-testid="inventory-create-paste-input"]').should("be.visible");
    cy.get('[data-testid="inventory-create-barcode-input"]').should("be.visible");
    cy.get('[data-testid="inventory-create-media-panel"]').should(
      "have.attr",
      "data-active-mode",
      "photo"
    );
    cy.get('[data-testid="inventory-item-editor-cancel"]').click();

    cy.get('[data-testid="inventory-paste-action"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and("contain", "Create Item")
      .and("not.contain", "Create Item From");
    cy.get('[data-testid="inventory-create-paste-input"]')
      .should("be.visible")
      .and("be.focused");
    cy.get('[data-testid="inventory-create-take-image"]')
      .scrollIntoView()
      .should("be.visible");
    cy.get('[data-testid="inventory-create-barcode-input"]').should("be.visible");
  });

  it("creates a draft item when any single field is provided", () => {
    const items: Array<{
      id: string;
      part_number: string;
      title: string;
      status: string;
      category: string;
      brand: string;
      priority: string;
      description: string;
    }> = [];

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsList");

    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body.part_number).to.match(/^DRAFT-/);
      expect(req.body).to.include({
        title: "Only Title Draft",
        brand: "Unknown",
        category: "General",
      });
      const created = {
        id: "item-title-only",
        part_number: req.body.part_number,
        title: req.body.title,
        status: "active",
        category: req.body.category,
        brand: req.body.brand,
        priority: "medium",
        description: req.body.description,
      };
      items.unshift(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createTitleOnlyItem");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-title"]').type("Only Title Draft");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@createTitleOnlyItem");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.contains("Only Title Draft").should("be.visible");
  });

  it("creates an item into an existing collection from the create modal", () => {
    const items: Array<{
      id: string;
      part_number: string;
      title: string;
      status: string;
      category: string;
      brand: string;
      priority: string;
      description: string;
    }> = [];

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsList");

    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body).to.include({
        part_number: "PN-STORE-1",
        title: "Store One Created Item",
      });
      const created = {
        id: "item-store-one-created",
        part_number: req.body.part_number,
        title: req.body.title,
        status: "active",
        category: req.body.category,
        brand: req.body.brand,
        priority: "medium",
        description: req.body.description,
      };
      items.unshift(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createStoreOneItem");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-collection"]')
      .should("be.visible")
      .clear()
      .type("Store 1");
    cy.get('[data-testid="inventory-item-part-number"]').type("PN-STORE-1");
    cy.get('[data-testid="inventory-item-title"]').type("Store One Created Item");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@createStoreOneItem");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 1");
    cy.get('[data-testid="inventory-collection-filter-selected"]').should(
      "contain",
      "Store 1"
    );
    cy.contains("Store One Created Item").should("be.visible");
    cy.window().then((win) => {
      expect(
        JSON.parse(
          win.localStorage.getItem("cabinet.inventory.item-folder-assignments.v1") ??
            "{}"
        )
      ).to.include({ "item-store-one-created": "Store 1" });
    });
  });

  it("creates a new collection from typed create modal collection text", () => {
    const items: Array<{
      id: string;
      part_number: string;
      title: string;
      status: string;
      category: string;
      brand: string;
      priority: string;
      description: string;
    }> = [];

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsList");

    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body).to.include({
        part_number: "PN-MODAL-SHELF",
        title: "Modal Shelf Item",
      });
      const created = {
        id: "item-modal-shelf",
        part_number: req.body.part_number,
        title: req.body.title,
        status: "active",
        category: req.body.category,
        brand: req.body.brand,
        priority: "medium",
        description: req.body.description,
      };
      items.unshift(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createModalShelfItem");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-collection"]').clear().type("Modal Shelf");
    cy.get('[data-testid="inventory-item-part-number"]').type("PN-MODAL-SHELF");
    cy.get('[data-testid="inventory-item-title"]').type("Modal Shelf Item");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@createModalShelfItem");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-active-context"]').should("contain", "Modal Shelf");
    cy.get('[data-testid="inventory-collection-filter-selected"]').should(
      "contain",
      "Modal Shelf"
    );
    cy.get('[data-testid="inventory-collection-filter-select"]').should(
      "contain",
      "Modal Shelf"
    );
    cy.contains("Modal Shelf Item").should("be.visible");
    cy.window().then((win) => {
      expect(
        JSON.parse(
          win.localStorage.getItem("cabinet.inventory.item-folder-assignments.v1") ??
            "{}"
        )
      ).to.include({ "item-modal-shelf": "Modal Shelf" });
    });
  });

  it("creates a draft item from only a captured image", () => {
    const items: Array<{
      id: string;
      part_number: string;
      title: string;
      status: string;
      category: string;
      brand: string;
      priority: string;
      description: string;
    }> = [];

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsList");

    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body.part_number).to.match(/^DRAFT-/);
      expect(req.body).to.include({
        title: "Photo draft item",
        brand: "Unknown",
        category: "General",
      });
      const created = {
        id: "item-photo-only",
        part_number: req.body.part_number,
        title: req.body.title,
        status: "active",
        category: req.body.category,
        brand: req.body.brand,
        priority: "medium",
        description: req.body.description,
      };
      items.unshift(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createPhotoOnlyItem");

    cy.intercept("POST", "/api/items/item-photo-only/photos", (req) => {
      req.reply({
        statusCode: 201,
        body: {
          id: "photo-draft-1",
          filename: "captured-draft.jpg",
          is_primary: true,
        },
      });
    }).as("uploadPhotoOnlyItem");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-create-take-image"]').should("be.visible");
    cy.get('[data-testid="inventory-create-photo-input"]').selectFile({
      contents: Cypress.Buffer.from("captured-photo-binary"),
      fileName: "captured-draft.jpg",
      mimeType: "image/jpeg",
    });
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@createPhotoOnlyItem");
    cy.wait("@uploadPhotoOnlyItem");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.contains("Photo draft item").should("be.visible");
  });

  it("keeps an empty draft item blocked until a field or image is provided", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsList");
    cy.intercept("POST", "/api/items").as("createItem");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.get('[data-testid="inventory-item-save-error"]')
      .should("be.visible")
      .and("contain", "Enter at least one field or add an image before saving.");
    cy.get("@createItem.all").should("have.length", 0);
  });
});
