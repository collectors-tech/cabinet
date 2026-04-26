describe("inventory-management", () => {
  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
  }

  it("renders inventory workspace, supports view toggle and filtering, and avoids 500", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-1",
            part_number: "PN-001",
            title: "Starter Item",
            status: "todo",
            category: "feature",
          },
          {
            id: "item-2",
            part_number: "PN-002",
            title: "Second Item",
            status: "used",
            category: "documentation",
          },
        ],
      },
    }).as("items");
    signIn();
    cy.wait("@items");
    cy.contains("500").should("not.exist");
    cy.contains("Oops! Something went wrong").should("not.exist");

    cy.contains("Inventory").should("be.visible");
    cy.contains("Collection Browser").should("be.visible");
    cy.contains("button", "New").should("be.visible");
    cy.contains("button", "Create").should("be.visible");

    cy.get('button[aria-label="Switch to cards view"]').click();
    cy.contains("Status:").should("be.visible");

    cy.get('button[aria-label="Switch to rows view"]')
      .click()
      .should("have.attr", "aria-pressed", "true");
    cy.get("table").should("be.visible");
    cy.contains("th", "Part #").should("be.visible");
    cy.contains("th", "Title").should("be.visible");
    cy.contains("th", "Condition").should("be.visible");
    cy.contains("th", "Category").should("be.visible");
    cy.contains("th", "Task").should("not.exist");
    cy.contains("th", "Priority").should("not.exist");
    cy.contains("PN-001").should("be.visible");
    cy.contains("todo").should("be.visible");
    cy.contains("feature").should("be.visible");

    cy.get('input[placeholder="Filter by title or part number..."]').type(
      "no-matching-task-xyz"
    );
    cy.contains("No results.").should("be.visible");
  });

  it("renders empty inventory state without global 500 fallback", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsEmpty");

    signIn();
    cy.wait("@itemsEmpty");

    cy.contains("500").should("not.exist");
    cy.contains("Oops! Something went wrong").should("not.exist");
    cy.contains("Inventory").should("be.visible");
    cy.contains("No results.").should("be.visible");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-002 shows inline error state and recovers on retry", () => {
    let attempts = 0;
    cy.intercept("GET", "/api/items", (req) => {
      attempts += 1;
      if (attempts === 1) {
        req.reply({
          statusCode: 500,
          body: { error: "failed_to_list_items" },
        });
        return;
      }
      req.reply({
        statusCode: 200,
        body: {
          items: [
            {
              id: "item-retry-1",
              part_number: "PN-RETRY-1",
              title: "Recovered Item",
              status: "todo",
              category: "feature",
            },
          ],
        },
      });
    }).as("itemsRetry");

    signIn();
    cy.wait("@itemsRetry");
    cy.get('[data-testid="inventory-load-error"]').should("be.visible");
    cy.contains("button", "Retry").click();
    cy.wait("@itemsRetry");
    cy.get('[data-testid="inventory-load-error"]').should("not.exist");
    cy.contains("Recovered Item").should("be.visible");
    cy.contains("500").should("not.exist");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-003 remains deterministic with bulk dataset filtering", () => {
    const bulk = Array.from({ length: 1200 }, (_, index) => ({
      id: `item-${index + 1}`,
      part_number: `PN-${index + 1}`,
      title: `Bulk Item ${index + 1}`,
      status: "todo",
      category: "feature",
    }));
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: bulk },
    }).as("itemsBulk");

    signIn();
    cy.wait("@itemsBulk");
    cy.contains("Items:").parent().contains("1200").should("be.visible");
    cy.contains("Page 1 of 120").should("exist");
    cy.get('button[aria-label="Switch to cards view"]').click();
    cy.contains("Status:").should("be.visible");
    cy.contains("button", "Rows").click();
    cy.get("table").should("be.visible");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-004 keeps summary compact in Collection Browser header and removes duplicate strips", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-compact-1",
            part_number: "PN-COMPACT-1",
            title: "Compact Layout Item",
            status: "todo",
            category: "feature",
          },
        ],
      },
    }).as("itemsCompact");

    signIn();
    cy.wait("@itemsCompact");

    cy.contains("Command Row").should("not.exist");
    cy.contains("Summary Strip").should("not.exist");

    cy.contains("Collection Browser")
      .closest('[data-slot="card"]')
      .within(() => {
        cy.contains(/Folders:\s*\d+/).should("be.visible");
        cy.contains(/Items:\s*\d+/).should("be.visible");
        cy.contains(/Active Brand:\s*\w+/).should("be.visible");
        cy.contains(/Active Category:\s*\w+/).should("be.visible");

        cy.contains(/Active Category:/)
          .should("be.visible")
          .then(($summary) => {
            const summaryTop = $summary[0].getBoundingClientRect().top;
            cy.get('input[placeholder="Filter by title or part number..."]')
              .should("be.visible")
              .then(($input) => {
                const inputTop = $input[0].getBoundingClientRect().top;
                expect(summaryTop).to.be.lessThan(inputTop);
              });
          });
      });
  });

  it("UI-SCREEN-INVENTORY-ITEMS-005 shows New action with adjacent Create menu", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsActions");

    signIn();
    cy.wait("@itemsActions");

    cy.get('[data-testid="inventory-new-action"]')
      .should("be.visible")
      .and("contain", "New")
      .click();
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-item-create-cancel"]').click();

    cy.get('[data-testid="inventory-create-menu-trigger"]')
      .should("be.visible")
      .and("contain", "Create")
      .click();
    cy.get('[data-testid="inventory-create-menu-item"]').should("be.visible").click();
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-create-menu-folder"]').should("not.exist");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-007 opens create-item workflow from toolbar", () => {
    const items: Array<{
      id: string;
      part_number: string;
      title: string;
      status: string;
      category: string;
    }> = [];
    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsCreate");
    cy.intercept("POST", "/api/items", (req) => {
      const created = {
        id: "item-inline-created",
        part_number: req.body.part_number,
        title: req.body.title,
        status: "active",
        category: "General",
      };
      items.push(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createToolbarItem");

    signIn();
    cy.wait("@itemsCreate");

    cy.get('[data-testid="inventory-create-menu-trigger"]').click();
    cy.get('[data-testid="inventory-create-menu-item"]').click();
    cy.get('[data-testid="inventory-item-title"]').type("Inline Created Item");
    cy.get('[data-testid="inventory-item-part-number"]').type("PN-CREATE-1");
    cy.get('[data-testid="inventory-item-create-submit"]').click();

    cy.wait("@createToolbarItem");
    cy.wait("@itemsCreate");
    cy.get('[data-testid="inventory-item-create-dialog"]').should("not.exist");
    cy.contains("Inline Created Item").should("be.visible");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-CREATE-1");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-007A creates a new item from the header Photos action", () => {
    const items: Array<{
      id: string;
      part_number: string;
      title: string;
      status: string;
      category: string;
    }> = [];
    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsPhotoCreate");
    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body).to.include({
        part_number: "PN-PHOTO-NEW",
        title: "Photo Created Item",
      });
      const created = {
        id: "item-photo-new",
        part_number: req.body.part_number,
        title: req.body.title,
        status: "active",
        category: "General",
      };
      items.push(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createPhotoItem");
    cy.intercept("POST", "/api/items/item-photo-new/photos", (req) => {
      req.reply({
        statusCode: 201,
        body: {
          id: "photo-new-1",
          filename: "new-inventory-photo.jpg",
          is_primary: true,
        },
      });
    }).as("uploadNewItemPhoto");

    signIn();
    cy.wait("@itemsPhotoCreate");

    cy.get('[data-testid="inventory-photos-action"]').click();
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-item-editor-mode"]').should(
      "contain",
      "Creating new item from photo"
    );
    cy.get('[data-testid="inventory-create-photo-input"]').selectFile({
      contents: Cypress.Buffer.from("new-item-photo-binary"),
      fileName: "new-inventory-photo.jpg",
      mimeType: "image/jpeg",
    });
    cy.get('[data-testid="inventory-item-part-number"]').type("PN-PHOTO-NEW");
    cy.get('[data-testid="inventory-item-title"]').type("Photo Created Item");
    cy.get('[data-testid="inventory-item-create-submit"]').click();

    cy.wait("@createPhotoItem");
    cy.wait("@uploadNewItemPhoto");
    cy.wait("@itemsPhotoCreate");
    cy.get('[data-testid="inventory-item-create-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-PHOTO-NEW");
    cy.contains("Photo Created Item").should("be.visible");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-007B creates a new item from the header Barcodes action", () => {
    const items: Array<{
      id: string;
      part_number: string;
      title: string;
      status: string;
      category: string;
    }> = [];
    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsBarcodeCreate");
    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body).to.include({
        part_number: "PN-BARCODE-NEW",
        title: "Barcode Created Item",
      });
      const created = {
        id: "item-barcode-new",
        part_number: req.body.part_number,
        title: req.body.title,
        status: "active",
        category: "General",
      };
      items.push(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createBarcodeItem");
    cy.intercept("POST", "/api/items/item-barcode-new/barcodes", (req) => {
      expect(req.body).to.deep.equal({ barcode: "012345678905" });
      req.reply({
        statusCode: 201,
        body: {
          id: "barcode-new-1",
          item_id: "item-barcode-new",
          barcode: "012345678905",
        },
      });
    }).as("attachNewItemBarcode");

    signIn();
    cy.wait("@itemsBarcodeCreate");

    cy.get('[data-testid="inventory-barcodes-action"]').click();
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-item-editor-mode"]').should(
      "contain",
      "Creating new item from barcode"
    );
    cy.get('[data-testid="inventory-create-barcode-input"]').type("012345678905");
    cy.get('[data-testid="inventory-item-part-number"]').type("PN-BARCODE-NEW");
    cy.get('[data-testid="inventory-item-title"]').type("Barcode Created Item");
    cy.get('[data-testid="inventory-item-create-submit"]').click();

    cy.wait("@createBarcodeItem");
    cy.wait("@attachNewItemBarcode");
    cy.wait("@itemsBarcodeCreate");
    cy.get('[data-testid="inventory-item-create-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-BARCODE-NEW");
    cy.contains("Barcode Created Item").should("be.visible");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-006 creates collection from the compact filter and auto-selects it", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-inline-1",
            part_number: "PN-INLINE-1",
            title: "Inline Collection Item",
            status: "todo",
            category: "feature",
          },
        ],
      },
    }).as("itemsInline");

    signIn();
    cy.wait("@itemsInline");

    cy.get('[data-testid="inventory-collection-add-root"]').click();
    cy.get('[data-testid="folder-tree-name-input"]').type("Inline Alpha");
    cy.get('[data-testid="folder-tree-create-submit"]').click();
    cy.get('[data-testid="inventory-collection-filter-selected"]').should(
      "contain",
      "Inline Alpha"
    );
  });

  it("UI-SCREEN-INVENTORY-ITEMS-009 persists create-edit save flow and keeps media attach usable", () => {
    const items = [
      {
        id: "item-existing-1",
        part_number: "PN-EXISTING-1",
        title: "Existing Inventory Item",
        status: "active",
        category: "Cars",
        brand: "AFX",
        priority: "medium",
        description: "Existing description",
      },
    ];
    const photos: Array<{ id: string; filename: string; is_primary: boolean }> = [];

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({
        statusCode: 200,
        body: { items },
      });
    }).as("itemsList");

    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body).to.include({
        part_number: "PN-CREATE-1",
        title: "Created Inventory Item",
        brand: "Tyco",
        category: "Cars",
      });
      const created = {
        id: "item-created-1",
        part_number: "PN-CREATE-1",
        title: "Created Inventory Item",
        status: "active",
        category: "Cars",
        brand: "Tyco",
        priority: "medium",
        description: "Freshly saved item",
      };
      items.unshift(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createItem");

    cy.intercept("PUT", "/api/items/item-created-1", (req) => {
      expect(req.body).to.include({
        title: "Created Inventory Item Updated",
        brand: "Aurora",
      });
      items[0] = {
        ...items[0],
        title: "Created Inventory Item Updated",
        brand: "Aurora",
      };
      req.reply({ statusCode: 200, body: items[0] });
    }).as("updateItem");

    cy.intercept("GET", "/api/items/item-created-1/photos", (req) => {
      req.reply({ statusCode: 200, body: { photos } });
    }).as("createdItemPhotos");

    cy.intercept("POST", "/api/items/item-created-1/photos", (req) => {
      photos.push({
        id: "photo-created-1",
        filename: "created-photo.jpg",
        is_primary: true,
      });
      req.reply({ statusCode: 201, body: photos[0] });
    }).as("uploadCreatedPhoto");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-editor-mode"]').should(
      "contain",
      "Creating new item draft"
    );
    cy.get('[data-testid="inventory-item-part-number"]').clear().type("PN-CREATE-1");
    cy.get('[data-testid="inventory-item-title"]').clear().type("Created Inventory Item");
    cy.get('[data-testid="inventory-item-brand"]').clear().type("Tyco");
    cy.get('[data-testid="inventory-item-category"]').clear().type("Cars");
    cy.get('[data-testid="inventory-item-description"]').clear().type("Freshly saved item");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@createItem");
    cy.wait("@itemsList");
    cy.wait("@createdItemPhotos");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-CREATE-1");
    cy.contains("Created Inventory Item").should("be.visible");

    cy.get(
      '[data-testid="inventory-item-row-item-created-1"] [data-testid="inventory-row-photos-action"]'
    ).click();
    cy.get('[data-testid="inventory-photos-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-photo-upload-input"]').selectFile({
      contents: Cypress.Buffer.from("created-photo-binary"),
      fileName: "created-photo.jpg",
      mimeType: "image/jpeg",
    });
    cy.wait("@uploadCreatedPhoto");
    cy.wait("@createdItemPhotos");
    cy.contains('[data-testid="inventory-photo-row"]', "created-photo.jpg").should(
      "be.visible"
    );
    cy.get('[data-testid="inventory-photos-dialog-close"]').click();
    cy.get('[data-testid="inventory-photos-dialog"]').should("not.exist");

    cy.get('[data-testid="inventory-item-row-item-created-1"] [data-testid="task-row-actions-trigger"]').click();
    cy.contains('[role="menuitem"]', "Edit").click();
    cy.get('[data-testid="inventory-item-editor-panel"]').should("be.visible");
    cy.get('[data-testid="inventory-item-title"]').clear().type("Created Inventory Item Updated");
    cy.get('[data-testid="inventory-item-brand"]').clear().type("Aurora");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@updateItem");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-panel"]').should("not.exist");
    cy.contains("Created Inventory Item Updated").should("be.visible");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-CREATE-1");
    cy.get(
      '[data-testid="inventory-item-row-item-created-1"] [data-testid="inventory-row-photos-action"]'
    ).click();
    cy.get('[data-testid="inventory-photos-dialog"]').should("be.visible");
    cy.contains('[data-testid="inventory-photo-row"]', "created-photo.jpg").should(
      "be.visible"
    );
  });

  it("UI-SCREEN-INVENTORY-ITEMS-010 saves and reapplies inventory views with real condition/category filters", () => {
    const items = [
      {
        id: "item-view-1",
        part_number: "PN-VIEW-001",
        title: "Road Alpha",
        status: "used",
        category: "Cars",
      },
      {
        id: "item-view-2",
        part_number: "PN-VIEW-002",
        title: "Road Zeta",
        status: "used",
        category: "Cars",
      },
      {
        id: "item-view-3",
        part_number: "PN-VIEW-003",
        title: "Road Bravo",
        status: "active",
        category: "Cars",
      },
      {
        id: "item-view-4",
        part_number: "PN-VIEW-004",
        title: "Plane Delta",
        status: "used",
        category: "Planes",
      },
    ];
    let profileSettings: Record<string, string> = {};
    let savedViewID = "";

    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "inventory-profile-1" },
    }).as("activeProfile");

    cy.intercept("GET", "/api/profiles/inventory-profile-1/settings", (req) => {
      req.reply({
        statusCode: 200,
        body: { settings: profileSettings },
      });
    }).as("inventoryProfileSettings");

    cy.intercept("PUT", "/api/profiles/inventory-profile-1/settings", (req) => {
      const body = typeof req.body === "string" ? JSON.parse(req.body) : req.body;
      profileSettings = body.settings ?? body ?? {};
      req.reply({
        statusCode: 200,
        body: { settings: profileSettings },
      });
    }).as("saveInventorySettings");

    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items },
    }).as("inventorySavedViewsItems");

    cy.clearLocalStorage("cabinet.viewMode.inventory");
    signIn();
    cy.wait("@activeProfile");
    cy.wait("@inventoryProfileSettings");
    cy.wait("@inventorySavedViewsItems");

    cy.get('button[aria-label="Switch to rows view"]')
      .click({ force: true })
      .should("have.attr", "aria-pressed", "true");
    cy.contains("button", "Condition").click();
    cy.contains('[cmdk-item]', "Used").click();
    cy.contains("button", "Category").click();
    cy.contains('[cmdk-item]', "Cars").click();
    cy.get('input[placeholder="Filter by title or part number..."]').type("Road");

    cy.contains("th", "Title").find("button").click();
    cy.contains('[role="menuitem"]', "Desc").click();

    cy.get('[data-testid="inventory-saved-view-save"]').click();
    cy.get('[data-testid="inventory-saved-view-name"]').type("Used Road Cars");
    cy.get('[data-testid="inventory-saved-view-submit"]').click();
    cy.wait("@saveInventorySettings")
      .then(() => {
        expect(profileSettings["inventory.saved-views.v1"]).to.contain("Used Road Cars");
        const views = JSON.parse(profileSettings["inventory.saved-views.v1"]);
        const savedView = views.find(
          (view: { name?: string }) => view.name === "Used Road Cars"
        );
        expect(savedView?.viewMode).to.equal("rows");
        savedViewID = savedView?.id ?? "";
        expect(savedViewID).to.not.equal("");
      });

    cy.contains("button", "Reset").click();
    cy.contains("button", "Cards").click();
    cy.contains("Road Bravo").should("be.visible");
    cy.contains("Plane Delta").should("be.visible");

    cy.reload();
    cy.wait("@activeProfile");
    cy.wait("@inventoryProfileSettings");
    cy.wait("@inventorySavedViewsItems");

    cy.then(() => {
      cy.get('[data-testid="inventory-saved-view-select"]').select(savedViewID);
    });

    cy.contains("button", "Rows").should("have.attr", "aria-pressed", "true");
    cy.get('input[placeholder="Filter by title or part number..."]').should(
      "have.value",
      "Road"
    );
    cy.contains("Road Zeta").should("be.visible");
    cy.contains("Road Alpha").should("be.visible");
    cy.contains("Road Bravo").should("not.exist");
    cy.contains("Plane Delta").should("not.exist");
    cy.get("table tbody tr").eq(0).should("contain", "PN-VIEW-002");
  });

});
