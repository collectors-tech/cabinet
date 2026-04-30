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
    cy.get('[data-testid="inventory-item-category-trigger"]').click();
    cy.get('[data-testid="inventory-item-category-search"]').type("Cars");
    cy.get('[data-testid="inventory-item-category-option-Cars"]').click();
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
      const viewportWidth = $panel[0].ownerDocument.defaultView?.innerWidth ?? Cypress.config("viewportWidth");
      expect(rect.left, "editor panel starts on right half").to.be.greaterThan(
        viewportWidth * 0.45
      );
      expect(rect.width, "editor panel keeps side-panel sizing").to.be.lessThan(
        viewportWidth
      );
    });
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Alpha Item");
    cy.get('[data-testid="inventory-item-editor-cancel"]').click();
    cy.get('[data-testid="inventory-item-editor-panel"]').should("not.exist");

    cy.get('[data-testid="inventory-item-row-item-alpha"] [data-testid="task-row-actions-trigger"]').click();
    cy.contains('[role="menuitem"]', "Edit").click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and("contain", "Edit Item");
    cy.get('[data-testid="inventory-item-editor-panel"]').should("not.exist");
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
    cy.get('[data-testid="inventory-item-editor-dialog"]').should("not.exist");
    cy.contains("Bravo Item Updated").should("be.visible");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-BRAVO");
  });

  it("selects multiple reusable categories and saves a newly typed category for future inventory edits", () => {
    const items = [
      {
        id: "item-category-alpha",
        part_number: "PN-CAT-ALPHA",
        title: "Category Alpha Item",
        status: "active",
        category: "Cars",
        brand: "AFX",
        priority: "medium",
        description: "Category alpha description",
      },
    ];

    cy.intercept("GET", "/api/profiles/*/settings", {
      statusCode: 200,
      body: {
        settings: {
          "inventory.category-options.v1": JSON.stringify([
            "General",
            "Cars",
            "Model Kit",
          ]),
        },
      },
    }).as("profileSettings");

    cy.intercept("PUT", "/api/profiles/*/settings", (req) => {
      const settings = req.body?.settings ?? {};
      const categories = JSON.parse(
        settings["inventory.category-options.v1"] ?? "[]"
      );
      if (categories.includes("Garage Kit")) {
        req.alias = "saveCategorySettings";
      }
      req.reply({ statusCode: 200, body: { settings } });
    });

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsList");

    cy.intercept("PUT", "/api/items/item-category-alpha", (req) => {
      expect(req.body.category).to.equal("Cars, Model Kit, Garage Kit");
      items[0] = { ...items[0], ...req.body };
      req.reply({ statusCode: 200, body: items[0] });
    }).as("updateCategoryAlpha");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-item-row-item-category-alpha"]').dblclick();
    cy.get('[data-testid="inventory-item-category"]').should("have.value", "Cars");
    cy.get('[data-testid="inventory-item-category-trigger"]').click();
    cy.get('[data-testid="inventory-item-category-option-Model Kit"]').click();
    cy.get('[data-testid="inventory-item-category-search"]').type("Garage Kit{enter}");
    cy.wait("@saveCategorySettings");
    cy.get('[data-testid="inventory-item-category"]')
      .should("have.value", "Cars, Model Kit, Garage Kit");
    cy.get('[data-testid="inventory-item-save"]').click();
    cy.wait("@updateCategoryAlpha");
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-item-row-item-category-alpha"] [data-testid="task-row-actions-trigger"]').click();
    cy.contains('[role="menuitem"]', "Edit").click();
    cy.get('[data-testid="inventory-item-category-trigger"]').click();
    cy.get('[data-testid="inventory-item-category-search"]').type("Garage Kit");
    cy.get('[data-testid="inventory-item-category-option-Garage Kit"]').should(
      "be.visible"
    );
  });

  it("uses a searchable create item category dropdown with enter-to-add", () => {
    cy.intercept("GET", "/api/profiles/*/settings", {
      statusCode: 200,
      body: {
        settings: {
          "inventory.category-options.v1": JSON.stringify([
            "General",
            "Cars",
            "Model Kit",
          ]),
        },
      },
    }).as("profileSettings");

    cy.intercept("PUT", "/api/profiles/*/settings", (req) => {
      const settings = req.body?.settings ?? {};
      const categories = JSON.parse(
        settings["inventory.category-options.v1"] ?? "[]"
      );
      if (categories.includes("Garage Kit")) {
        req.alias = "saveNewCategoryOption";
      }
      req.reply({ statusCode: 200, body: { settings } });
    });

    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsList");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-category-trigger"]').click();
    cy.get('[data-testid="inventory-item-category-search"]')
      .type("Model");
    cy.get('[data-testid="inventory-item-category-option-Model Kit"]').click();
    cy.get('[data-testid="inventory-item-category"]').should(
      "have.value",
      "Model Kit"
    );

    cy.get('[data-testid="inventory-item-category-search"]')
      .clear()
      .type("Garage Kit");
    cy.get('[data-testid="inventory-item-category-add-hint"]')
      .should("be.visible")
      .and("contain", "Press Enter to add");
    cy.get('[data-testid="inventory-item-category-search"]').type("{enter}");
    cy.wait("@saveNewCategoryOption");
    cy.get('[data-testid="inventory-item-category"]').should(
      "have.value",
      "Model Kit, Garage Kit"
    );
  });

  it("shows gallery, evidence, pricing, tags, urls, description, notes, and navigation in the item edit panel", () => {
    const items = [
      {
        id: "item-alpha",
        part_number: "PN-ALPHA",
        title: "Alpha Item",
        status: "active",
        category: "Cars",
        brand: "AFX",
        priority: "medium",
        description: "Alpha public description",
        notes: "Alpha private item notes",
        tags: ["sealed", "rare"],
        source_urls: ["https://example.test/source-alpha"],
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
        notes: "Bravo private notes",
        tags: ["loose"],
        source_urls: [],
      },
    ];

    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items },
    }).as("itemsList");

    cy.intercept("GET", "/api/items/item-alpha/photos", {
      statusCode: 200,
      body: {
        photos: [
          { id: "alpha-front", filename: "alpha-front.jpg", is_primary: true },
          { id: "alpha-back", filename: "alpha-back.jpg", is_primary: false },
        ],
      },
    }).as("alphaPhotos");
    cy.intercept("POST", "/api/items/item-alpha/photos", (req) => {
      expect(req.headers["content-type"]).to.contain("multipart/form-data");
      req.reply({
        statusCode: 201,
        body: { id: "alpha-side", filename: "alpha-side.jpg", is_primary: false },
      });
    }).as("addAlphaPhoto");
    cy.intercept("DELETE", "/api/items/item-alpha/photos/alpha-back", {
      statusCode: 204,
      body: {},
    }).as("deleteAlphaBackPhoto");
    cy.intercept("PUT", "/api/items/item-alpha/photos/alpha-back/rotate", (req) => {
      expect(req.body).to.deep.equal({ direction: "left" });
      req.reply({ statusCode: 200, body: {} });
    }).as("rotateAlphaBackPhotoLeft");
    cy.intercept("PUT", "/api/items/item-alpha/photos/alpha-front/rotate", (req) => {
      expect(req.body).to.deep.equal({ direction: "right" });
      req.reply({ statusCode: 200, body: {} });
    }).as("rotateAlphaFrontPhotoRight");
    cy.intercept("GET", "/api/items/item-alpha/photos/*/file?*", {
      statusCode: 200,
      headers: { "content-type": "image/gif" },
      body: "R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw==",
    });
    cy.intercept("GET", "/api/items/item-alpha/barcodes", {
      statusCode: 200,
      body: {
        barcodes: [
          {
            id: "barcode-alpha",
            item_id: "item-alpha",
            barcode: "1234567890123",
            created_at: "2026-04-28T01:00:00Z",
          },
        ],
      },
    }).as("alphaBarcodes");
    cy.intercept("GET", "/api/items/item-alpha/instances", {
      statusCode: 200,
      body: {
        instances: [
          {
            id: "instance-alpha",
            item_id: "item-alpha",
            condition: "sealed",
            status: "sealed",
            quantity: 2,
            storage_location: "Shelf A",
            acquisition_price: 49.95,
            acquisition_date: "2026-04-21",
            notes: "Shelf note",
            created_at: "2026-04-21T00:00:00Z",
            updated_at: "2026-04-22T00:00:00Z",
          },
        ],
      },
    }).as("alphaInstances");
    cy.intercept("GET", "/api/pricing/history?item_id=item-alpha", {
      statusCode: 200,
      body: {
        history: [
          {
            snapshot_date: "2026-04-24",
            source: "ebay",
            min_price: 44,
            median_price: 50,
            latest_price: 48,
            stock_count: 5,
          },
          {
            snapshot_date: "2026-04-25",
            source: "ebay",
            min_price: 46,
            median_price: 54,
            latest_price: 57,
            stock_count: 4,
          },
          {
            snapshot_date: "2026-04-26",
            source: "tcgplayer",
            min_price: 49,
            median_price: 58,
            latest_price: 61,
            stock_count: 6,
          },
          {
            snapshot_date: "2026-04-27",
            source: "ebay",
            min_price: 52,
            median_price: 60,
            latest_price: 64,
            stock_count: 3,
          },
        ],
      },
    }).as("alphaPricing");

    signIn();
    cy.wait("@itemsList");
    cy.get('[data-testid="inventory-item-row-item-alpha"]').dblclick();

    cy.wait("@alphaPhotos");
    cy.wait("@alphaBarcodes");
    cy.wait("@alphaInstances");
    cy.wait("@alphaPricing");

    cy.get('[data-testid="inventory-item-editor-panel"]')
      .should("be.visible")
      .and("contain", "Edit Item");
    cy.get('[data-testid="inventory-item-gallery"]').should("be.visible");
    cy.get('[data-testid="inventory-item-gallery-preview"]')
      .should("be.visible")
      .and("have.attr", "alt", "alpha-front.jpg");
    cy.get('[data-testid="inventory-item-gallery-thumb"]').should("have.length", 2);
    cy.get('[data-testid="inventory-item-gallery-add"]').should("be.visible");
    cy.get('[data-testid="inventory-item-gallery-add-input"]').selectFile(
      {
        contents: Cypress.Buffer.from("fake-image"),
        fileName: "alpha-side.jpg",
        mimeType: "image/jpeg",
      },
      { force: true }
    );
    cy.wait("@addAlphaPhoto");
    cy.wait("@alphaPhotos");
    cy.get('[data-testid="inventory-item-gallery-preview-rotate-left"]')
      .should("be.visible")
      .and("have.attr", "aria-label", "Rotate alpha-front.jpg left");
    cy.get('[data-testid="inventory-item-gallery-preview-rotate-right"]')
      .should("be.visible")
      .and("have.attr", "aria-label", "Rotate alpha-front.jpg right")
      .click();
    cy.wait("@rotateAlphaFrontPhotoRight");
    cy.wait("@alphaPhotos");
    cy.get('[data-testid="inventory-item-gallery-thumb"]').eq(1).click();
    cy.get('[data-testid="inventory-item-gallery-preview"]').should(
      "have.attr",
      "alt",
      "alpha-back.jpg"
    );
    cy.get('[data-testid="inventory-item-gallery-open"]').click();
    cy.get('[data-testid="inventory-photo-fullscreen"]')
      .should("be.visible")
      .and("contain", "alpha-back.jpg");
    cy.get('[data-testid="inventory-photo-prev"]').click();
    cy.get('[data-testid="inventory-photo-fullscreen"]').should(
      "contain",
      "alpha-front.jpg"
    );
    cy.get('[data-testid="inventory-photo-fullscreen-close"]').click();

    cy.get('[data-testid="inventory-item-gallery-preview-rotate-left"]').click();
    cy.wait("@rotateAlphaBackPhotoLeft");
    cy.wait("@alphaPhotos");
    cy.get('[data-testid="inventory-item-gallery-thumb-delete"]')
      .eq(1)
      .should("be.visible")
      .and("have.attr", "aria-label", "Delete alpha-back.jpg")
      .click();
    cy.wait("@deleteAlphaBackPhoto");
    cy.wait("@alphaPhotos");

    cy.get('[data-testid="inventory-item-tags"]').should(
      "have.value",
      "sealed, rare"
    );
    cy.get('[data-testid="inventory-item-source-url-0"]')
      .scrollIntoView()
      .should("be.visible")
      .and("have.attr", "href", "https://example.test/source-alpha");
    cy.get('[data-testid="inventory-item-pricing-panel"]')
      .scrollIntoView()
      .should("contain", "$49.95")
      .and("contain", "$64.00")
      .and("contain", "ebay");
    cy.get('[data-testid="inventory-item-price-history-chart"]')
      .should("be.visible")
      .and("contain", "Price history")
      .and("contain", "$48.00")
      .and("contain", "$64.00")
      .find("svg polyline")
      .should("have.attr", "points")
      .and("match", /,/);
    cy.get('[data-testid="inventory-item-price-history-chart"]').should(
      "contain",
      "4 price points"
    );
    cy.get('[data-testid="inventory-item-barcodes-panel"]').should(
      "contain",
      "1234567890123"
    );
    cy.get('[data-testid="inventory-item-notes"]').should(
      "have.value",
      "Alpha private item notes"
    );
    cy.get('[data-testid="inventory-item-instance-notes"]').should(
      "contain",
      "Shelf note"
    );
    cy.get('[data-testid="inventory-item-description"]').should(
      "have.value",
      "Alpha public description"
    );
    cy.get('[data-testid="inventory-item-editor-next"]').click();
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Bravo Item");
  });

  it("edits existing item pricing evidence and creates default instance evidence from the right panel", () => {
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
        notes: "Alpha notes",
        tags: [],
        source_urls: [],
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
        notes: "Bravo notes",
        tags: [],
        source_urls: [],
      },
    ];
    let alphaInstances = [
      {
        id: "instance-alpha",
        item_id: "item-alpha",
        condition: "used",
        status: "loose",
        quantity: 2,
        storage_location: "Shelf A",
        acquisition_price: 49.95,
        acquisition_date: "2026-04-21",
        notes: "Shelf note",
      },
    ];
    let bravoInstances: Array<{
      id: string;
      item_id: string;
      condition: string;
      status: string;
      quantity: number;
      storage_location: string;
      acquisition_price: number;
      acquisition_date: string;
      notes: string;
    }> = [];

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsList");
    cy.intercept("GET", "/api/items/*/photos", {
      statusCode: 200,
      body: { photos: [] },
    });
    cy.intercept("GET", "/api/items/*/barcodes", {
      statusCode: 200,
      body: { barcodes: [] },
    });
    cy.intercept("GET", "/api/pricing/history?item_id=*", {
      statusCode: 200,
      body: { history: [] },
    });
    cy.intercept("GET", "/api/items/item-alpha/instances", (req) => {
      req.reply({ statusCode: 200, body: { instances: alphaInstances } });
    }).as("alphaInstances");
    cy.intercept("GET", "/api/items/item-bravo/instances", (req) => {
      req.reply({ statusCode: 200, body: { instances: bravoInstances } });
    }).as("bravoInstances");
    cy.intercept("PUT", "/api/items/item-alpha", (req) => {
      req.reply({ statusCode: 200, body: { ...items[0], ...req.body } });
    }).as("updateAlphaItem");
    cy.intercept("PUT", "/api/items/item-bravo", (req) => {
      req.reply({ statusCode: 200, body: { ...items[1], ...req.body } });
    }).as("updateBravoItem");
    cy.intercept("PUT", "/api/items/item-alpha/instances/instance-alpha", (req) => {
      expect(req.body).to.include({
        condition: "8 - Like new",
        status: "sealed",
        quantity: 3,
        storage_location: "Vault B",
        acquisition_price: 75.5,
        acquisition_date: "2026-04-22",
        notes: "Updated evidence note",
      });
      alphaInstances = [{ ...alphaInstances[0], ...req.body }];
      req.reply({ statusCode: 200, body: alphaInstances[0] });
    }).as("updateAlphaInstance");
    cy.intercept("POST", "/api/items/item-bravo/instances", (req) => {
      expect(req.body).to.include({
        condition: "10+ - New, in packaging",
        status: "loose",
        quantity: 1,
        storage_location: "Case C",
        acquisition_price: 22,
        acquisition_date: "2026-04-23",
        notes: "New instance evidence",
      });
      const created = {
        id: "instance-bravo-created",
        item_id: "item-bravo",
        ...req.body,
      };
      bravoInstances = [created];
      req.reply({ statusCode: 201, body: created });
    }).as("createBravoInstance");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-item-row-item-alpha"]').dblclick();
    cy.wait("@alphaInstances");
    cy.get('[data-testid="inventory-instance-price"]').clear().type("75.50");
    cy.get('[data-testid="inventory-instance-quantity"]').clear().type("3");
    cy.get('[data-testid="inventory-instance-condition"]').select("8 - Like new");
    cy.get('[data-testid="inventory-instance-status"]').clear().type("sealed");
    cy.get('[data-testid="inventory-instance-storage-location"]')
      .clear()
      .type("Vault B");
    cy.get('[data-testid="inventory-instance-acquisition-date"]')
      .clear()
      .type("2026-04-22");
    cy.get('[data-testid="inventory-instance-notes-field"]')
      .clear()
      .type("Updated evidence note");
    cy.get('[data-testid="inventory-item-save"]').click();
    cy.wait("@updateAlphaItem");
    cy.wait("@updateAlphaInstance");

    cy.get('[data-testid="inventory-item-row-item-bravo"]').dblclick();
    cy.wait("@bravoInstances");
    cy.get('[data-testid="inventory-instance-price"]').clear().type("22");
    cy.get('[data-testid="inventory-instance-quantity"]').clear().type("1");
    cy.get('[data-testid="inventory-instance-condition"]').select(
      "10+ - New, in packaging"
    );
    cy.get('[data-testid="inventory-instance-status"]').clear().type("loose");
    cy.get('[data-testid="inventory-instance-storage-location"]')
      .clear()
      .type("Case C");
    cy.get('[data-testid="inventory-instance-acquisition-date"]')
      .clear()
      .type("2026-04-23");
    cy.get('[data-testid="inventory-instance-notes-field"]')
      .clear()
      .type("New instance evidence");
    cy.get('[data-testid="inventory-item-save"]').click();
    cy.wait("@updateBravoItem");
    cy.wait("@createBravoInstance");
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

  it("opens the create item modal at a wide desktop size", () => {
    cy.viewport(1440, 900);
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsList");

    signIn();
    cy.wait("@itemsList");

    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and(($dialog) => {
        expect($dialog[0].getBoundingClientRect().width).to.be.greaterThan(1000);
      });
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
    cy.get('[data-testid="inventory-active-folder-label"]').should(
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
    cy.get('[data-testid="inventory-active-folder-label"]').should(
      "contain",
      "Modal Shelf"
    );
    cy.get('[data-testid="inventory-folder-tree"]').should("contain", "Modal Shelf");
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

  it("opens the image picker when creating an item from the header image action", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsList");

    signIn();
    cy.wait("@itemsList");

    cy.window().then((win) => {
      cy.spy(win.HTMLElement.prototype, "click").as("elementClick");
    });

    cy.get('[data-testid="inventory-photos-action"]').click();
    cy.get('[data-testid="inventory-item-editor-dialog"]')
      .should("be.visible")
      .and("contain", "Create Item");
    cy.get('[data-testid="inventory-create-media-panel"]').should(
      "have.attr",
      "data-active-mode",
      "photo"
    );
    cy.get('[data-testid="inventory-create-photo-input"]').should("exist");
    cy.get("@elementClick").should("have.been.called");
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
