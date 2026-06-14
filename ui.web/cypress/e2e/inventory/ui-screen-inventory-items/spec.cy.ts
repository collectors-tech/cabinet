describe("inventory-management", () => {
  function signIn() {
    cy.viewport(1280, 900);
    cy.e2eReset();
    cy.clearLocalStorage("cabinet.viewMode.inventory");
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
  }

  function setInventoryItemCategory(category: string) {
    cy.get('[data-testid="inventory-item-category-trigger"]').click();
    cy.get(`[data-testid="inventory-item-category-option-${category}"]`).click();
    cy.get('[data-testid="inventory-item-category"]').should("have.value", category);
  }

  function scrollInventoryTable(position: "left" | "right") {
    cy.get(
      '[data-testid="inventory-table-surface"] [data-slot="table-container"]'
    ).scrollTo(position);
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
    cy.contains("Collection Browser").should("not.exist");
    cy.get('[data-testid="inventory-new-action"]').should("be.visible");
    cy.get('[data-testid="inventory-create-menu-trigger"]').should("be.visible");

    cy.get('button[aria-label="Switch to cards view"]').click();
    cy.contains("Status:").should("be.visible");

    cy.get('button[aria-label="Switch to rows view"]')
      .click()
      .should("have.attr", "aria-pressed", "true");
    cy.get("table").should("be.visible");
    cy.contains("th", "Part #").should("exist");
    cy.contains("th", "Title").should("exist");
    cy.contains("th", "Condition").should("exist");
    cy.contains("th", "Category").should("exist");
    cy.contains("th", "Task").should("not.exist");
    cy.contains("th", "Priority").should("not.exist");
    cy.contains("PN-001").should("be.visible");
    cy.contains("todo").should("exist");
    cy.contains("feature").should("exist");

    cy.get('input[placeholder^="Filter by title"]').type(
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

  it("UI-SCREEN-INVENTORY-ITEMS-004 keeps summary compact without duplicate strips or redundant section heading", () => {
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
    cy.contains("Collection Browser").should("not.exist");

    cy.contains(/Folders:\s*\d+/)
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
            cy.get('input[placeholder^="Filter by title"]')
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
      .and("have.attr", "aria-label", "New item")
      .and("not.contain", "New")
      .click();
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-item-create-cancel"]').click();

    cy.get('[data-testid="inventory-create-menu-trigger"]')
      .should("be.visible")
      .and("have.attr", "aria-label", "Create menu")
      .and("not.contain", "Create")
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

  it("UI-SCREEN-INVENTORY-ITEMS-006 creates collection from the folder tree and auto-selects it", () => {
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

    cy.get('[data-testid="folder-tree-add-root"]')
      .should("have.attr", "aria-label", "Add root folder")
      .and("not.contain", "New Collection");
    cy.get('[data-testid="folder-tree-add-root"]').click();
    cy.get('[data-testid="folder-tree-name-input"]').type("Inline Alpha");
    cy.get('[data-testid="folder-tree-create-submit"]').click();
    cy.get('[data-testid="collection-active-context"]').should("contain", "Inline Alpha");
    cy.get('[data-testid="folder-tree-item-inline-alpha"]')
      .scrollIntoView()
      .should("be.visible");
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
    setInventoryItemCategory("Cars");
    cy.get('[data-testid="inventory-item-description"]').clear().type("Freshly saved item");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@createItem");
    cy.wait("@itemsList");
    cy.wait("@createdItemPhotos");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-CREATE-1");
    scrollInventoryTable("left");
    cy.contains("Created Inventory Item").should("be.visible");

    scrollInventoryTable("right");
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

    scrollInventoryTable("right");
    cy.get('[data-testid="inventory-item-row-item-created-1"] [data-testid="task-row-actions-trigger"]').click();
    cy.contains('[role="menuitem"]', "Edit").click();
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-item-title"]').clear().type("Created Inventory Item Updated");
    cy.get('[data-testid="inventory-item-brand"]').clear().type("Aurora");
    cy.get('[data-testid="inventory-item-save"]').click();

    cy.wait("@updateItem");
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    scrollInventoryTable("left");
    cy.contains("Created Inventory Item Updated").should("be.visible");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-CREATE-1");
    scrollInventoryTable("right");
    cy.get(
      '[data-testid="inventory-item-row-item-created-1"] [data-testid="inventory-row-photos-action"]'
    ).click();
    cy.get('[data-testid="inventory-photos-dialog"]').should("be.visible");
    cy.contains('[data-testid="inventory-photo-row"]', "created-photo.jpg").should(
      "be.visible"
    );
  });

  it("UI-SCREEN-INVENTORY-ITEMS-010 saves and reapplies inventory views with search, sorting, and row mode", () => {
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
    cy.get('input[placeholder^="Filter by title"]').type("Road");

    cy.contains("th", "Title").find("button").click();
    cy.contains('[role="menuitem"]', "Desc").click();

    cy.get('[data-testid="inventory-saved-view-save"]').click();
    cy.get('[data-testid="inventory-saved-view-name"]').type("Road Items");
    cy.get('[data-testid="inventory-saved-view-submit"]').click();
    cy.wait("@saveInventorySettings")
      .then(() => {
        expect(profileSettings["inventory.saved-views.v1"]).to.contain("Road Items");
        const views = JSON.parse(profileSettings["inventory.saved-views.v1"]);
        const savedView = views.find(
          (view: { name?: string }) => view.name === "Road Items"
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
    cy.get('input[placeholder^="Filter by title"]').should(
      "have.value",
      "Road"
    );
    cy.contains("Road Zeta").should("be.visible");
    cy.contains("Road Alpha").should("be.visible");
    cy.contains("Road Bravo").should("be.visible");
    cy.contains("Plane Delta").should("not.exist");
    cy.get("table tbody tr").eq(0).should("contain", "PN-VIEW-002");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-011 scopes condition choices by item type and restores compact filters", () => {
    cy.intercept("GET", "/api/inventory/grading/enums", {
      statusCode: 200,
      body: {
        condition_grades: ["M", "NM"],
        packaging_grades: ["sealed"],
        item_type_condition_scales: [
          {
            item_type: "Slot Cars",
            conditions: ["10+ - New, in packaging", "8 - Like new"],
          },
          {
            item_type: "Trading Cards",
            conditions: ["Mint (M)", "Near Mint (NM)", "Played (PL)"],
          },
        ],
      },
    }).as("gradingEnums");
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-slot-1",
            part_number: "PN-SLOT-1",
            title: "Slot Car Item",
            status: "10+ - New, in packaging",
            category: "Slot Car",
            item_type: "Slot Cars",
          },
          {
            id: "item-card-1",
            part_number: "PN-CARD-1",
            title: "Trading Card Item",
            status: "Near Mint (NM)",
            category: "Trading Card",
            item_type: "Trading Cards",
          },
        ],
      },
    }).as("itemsScopedConditions");

    signIn();
    cy.wait("@gradingEnums");
    cy.wait("@itemsScopedConditions");

    cy.get('[data-testid="inventory-table-toolbar"]').within(() => {
      cy.contains("button", "Condition").should("be.visible");
      cy.contains("button", "Category").should("be.visible");
    });
    cy.contains("button", "Category").click();
    cy.contains('[role="option"]', "Trading Card").click();
    cy.contains("Trading Card Item").should("be.visible");
    cy.contains("Slot Car Item").should("not.exist");
    cy.get("body").type("{esc}");
    cy.contains("button", "Reset").click();

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-type"]').scrollIntoView().should("be.visible");
    cy.get('[data-testid="inventory-item-type"]').select("Trading Cards");
    cy.get('[data-testid="inventory-instance-condition"]').select("Near Mint (NM)");
    cy.get('[data-testid="inventory-instance-condition"]').should(
      "have.value",
      "Near Mint (NM)"
    );
    cy.get('[data-testid="inventory-instance-condition"]')
      .find("option")
      .should("contain", "Played (PL)")
      .and("not.contain", "10+ - New, in packaging");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-012 keeps dense row columns readable", () => {
    cy.viewport(1280, 900);
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-dense-1",
            part_number: "PN-DENSE-0000000001",
            title: "Dense Inventory Item With A Representative Long Display Title",
            status: "10+ - New, in packaging",
            category: "Slot Cars, Limited Edition",
            item_type: "Slot Cars",
            packaging_grade_type: "Factory sealed long card",
          },
          {
            id: "item-dense-2",
            part_number: "PN-DENSE-0000000002",
            title: "Second Dense Inventory Item",
            status: "8 - Like new",
            category: "Trading Cards",
            item_type: "Trading Cards",
            packaging_grade_type: "Sleeved and boxed",
          },
        ],
      },
    }).as("itemsDenseRows");

    signIn();
    cy.wait("@itemsDenseRows");

    cy.get('button[aria-label="Switch to rows view"]')
      .click({ force: true })
      .should("have.attr", "aria-pressed", "true");

    cy.get(
      '[data-testid="inventory-table-surface"] [data-slot="table-container"]'
    ).then(($surface) => {
      expect($surface[0].scrollWidth).to.be.greaterThan(
        $surface[0].clientWidth
      );
    });

    cy.contains("th", "Part #").should("be.visible");
    cy.contains("th", "Title").should("be.visible");
    cy.contains("th", "Condition").should("exist");

    cy.get(
      '[data-testid="inventory-table-surface"] [data-slot="table-container"]'
    ).scrollTo("right");
    cy.contains("th", "Packaging").should("be.visible");
    cy.contains("th", "Category").should("be.visible");
    cy.get(
      '[data-testid="inventory-item-row-item-dense-1"] [data-testid="inventory-row-packaging-grade"]'
    )
      .should("be.visible")
      .and("have.attr", "title", "Factory sealed long card");
    cy.get(
      '[data-testid="inventory-item-row-item-dense-1"] [data-testid="inventory-row-photos-action"]'
    ).should("be.visible");

    cy.contains("th", "Item type").then(($itemType) => {
      cy.contains("th", "Packaging").then(($packaging) => {
        cy.contains("th", "Category").then(($category) => {
          const itemTypeRight = $itemType[0].getBoundingClientRect().right;
          const packagingLeft = $packaging[0].getBoundingClientRect().left;
          const packagingRight = $packaging[0].getBoundingClientRect().right;
          const categoryLeft = $category[0].getBoundingClientRect().left;

          expect(itemTypeRight).to.be.at.most(packagingLeft);
          expect(packagingRight).to.be.at.most(categoryLeft);
        });
      });
    });
  });

  it("UI-SCREEN-INVENTORY-ITEMS-001/008 covers search, filters, sort, reset, and bulk selection", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-table-1",
            part_number: "PN-TABLE-001",
            title: "Road Alpha",
            status: "used",
            category: "Cars",
            item_type: "Slot Cars",
            packaging_grade_type: "Carded",
          },
          {
            id: "item-table-2",
            part_number: "PN-TABLE-002",
            title: "Road Zeta",
            status: "used",
            category: "Cars",
            item_type: "Slot Cars",
            packaging_grade_type: "Loose",
          },
          {
            id: "item-table-3",
            part_number: "PN-TABLE-003",
            title: "Road Bravo",
            status: "active",
            category: "Cars",
            item_type: "Slot Cars",
            packaging_grade_type: "Carded",
          },
          {
            id: "item-table-4",
            part_number: "PN-TABLE-004",
            title: "Plane Delta",
            status: "used",
            category: "Planes",
            item_type: "Model Planes",
            packaging_grade_type: "Boxed",
          },
        ],
      },
    }).as("itemsTableControls");

    signIn();
    cy.wait("@itemsTableControls");

    cy.get('button[aria-label="Switch to rows view"]')
      .click({ force: true })
      .should("have.attr", "aria-pressed", "true");

    cy.get('[data-testid="inventory-table-search-input"]').type("Road");
    cy.contains("Road Alpha").should("be.visible");
    cy.contains("Road Zeta").should("be.visible");
    cy.contains("Road Bravo").should("be.visible");
    cy.contains("Plane Delta").should("not.exist");

    cy.get('[data-testid="inventory-table-toolbar"]')
      .contains("button", "Condition")
      .click();
    cy.contains('[role="option"]', "used").click();
    cy.get("body").type("{esc}");
    cy.contains("Road Alpha").should("be.visible");
    cy.contains("Road Zeta").should("be.visible");
    cy.contains("Road Bravo").should("not.exist");

    cy.get('[data-testid="inventory-table-toolbar"]')
      .contains("button", "Category")
      .click();
    cy.contains('[role="option"]', "Cars").click();
    cy.get("body").type("{esc}");
    cy.contains("Road Alpha").should("be.visible");
    cy.contains("Road Zeta").should("be.visible");
    cy.contains("Plane Delta").should("not.exist");

    cy.contains("th", "Title").find("button").click();
    cy.contains('[role="menuitem"]', "Desc").click();
    cy.get("table tbody tr").first().should("contain", "PN-TABLE-002");

    cy.contains("button", "Reset").click();
    cy.get('[data-testid="inventory-table-search-input"]').should("have.value", "");
    cy.contains("Road Bravo").should("be.visible");
    cy.contains("Plane Delta").should("be.visible");

    cy.get('[data-testid="inventory-item-row-item-table-1"]')
      .find('button[role="checkbox"][aria-label="Select row"]')
      .click();
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.get('[data-testid="inventory-item-row-item-table-1"]').should(
      "have.attr",
      "data-state",
      "selected"
    );
    cy.get('[role="toolbar"][aria-label^="Bulk actions for 1 selected"]')
      .should("be.visible")
      .within(() => {
        cy.contains("selected").should("be.visible");
        cy.get('button[aria-label="Update status"]').should("be.visible");
        cy.get('button[aria-label="Update priority"]').should("be.visible");
        cy.get('button[aria-label="Export tasks"]').should("be.visible");
        cy.get('button[aria-label="Delete selected tasks"]').should("be.visible");
      });

    cy.get('[role="toolbar"][aria-label^="Bulk actions for 1 selected"]')
      .focus()
      .type("{esc}");
    cy.get('[role="toolbar"][aria-label^="Bulk actions"]').should("not.exist");
    cy.get('[data-testid="inventory-item-row-item-table-1"]').should(
      "not.have.attr",
      "data-state",
      "selected"
    );

    cy.get('button[role="checkbox"][aria-label="Select all"]').click();
    cy.get('[role="toolbar"][aria-label^="Bulk actions for 4 selected"]').should(
      "be.visible"
    );
    cy.get('button[aria-label="Clear selection"]').click();
    cy.get('[role="toolbar"][aria-label^="Bulk actions"]').should("not.exist");
  });

});
