describe("inventory Bonza product URL ingest", () => {
  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
  }

  it("UI-SCREEN-INVENTORY-BONZA-001 processes a Bonza pasted URL into a confirm-before-create item draft", () => {
    const items: Array<Record<string, unknown>> = [];
    const bonzaURL = "https://bonzaslotcars.com.au/product/bonza-mug-white/";

    cy.intercept("GET", "/api/items", (req) => {
      req.reply({ statusCode: 200, body: { items } });
    }).as("itemsBonzaPaste");

    cy.intercept("POST", "/api/providers/product-url/ingest", (req) => {
      expect(req.body).to.deep.eq({ url: bonzaURL, capture_for_review: true });
      req.reply({
        statusCode: 200,
        body: {
          mode: "provider_product_url_ingest",
          provider: "bonzaslotcars",
          family: "woocommerce",
          draft: {
            provider_product_id: "19603",
            title: "BONZA MUG WHITE",
            source_url: bonzaURL,
            description: "White Bonza branded mug.",
            price: 9.95,
            currency: "AUD",
            stock_state: "in_stock",
            stock_count: 3,
            categories: ["AFX ACCESSORIES HO", "MERCHANDISE"],
            attributes: { Brand: "AFX", Scale: "1:64", Type: "Tracks" },
            image_urls: [
              "https://bonzaslotcars.com.au/wp-content/uploads/BONZA-MUG.jpg",
            ],
          },
          evidence: {
            provider: "bonzaslotcars",
            family: "woocommerce",
            extraction_method: "store_api",
            provider_product_id: "19603",
            original_url: bonzaURL,
            normalized_url: bonzaURL,
            observed_at: "2026-05-29T15:00:00Z",
            source_summary: "WooCommerce Store API product detail",
          },
          duplicates: [],
        },
      });
    }).as("bonzaProductIngest");

    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body).to.include({
        part_number: "BONZA-19603",
        title: "BONZA MUG WHITE",
        brand: "AFX",
        category: "AFX ACCESSORIES HO, MERCHANDISE",
      });
      expect(req.body.source_urls).to.deep.eq([bonzaURL]);
      expect(req.body.notes).to.contain("Provider product ID: 19603");
      expect(req.body.description).to.contain("BONZA-MUG.jpg");
      const created = {
        id: "item-bonza-mug",
        part_number: "BONZA-19603",
        title: "BONZA MUG WHITE",
        status: "active",
        category: "AFX ACCESSORIES HO, MERCHANDISE",
        brand: "AFX",
        source_urls: [bonzaURL],
      };
      items.unshift(created);
      req.reply({ statusCode: 201, body: created });
    }).as("createBonzaItem");

    signIn();
    cy.wait("@itemsBonzaPaste");
    cy.window().then((win) => {
      if (!win.navigator.clipboard) {
        Object.defineProperty(win.navigator, "clipboard", {
          value: { readText: () => Promise.resolve("") },
          configurable: true,
        });
      }
      cy.stub(win.navigator.clipboard, "readText").resolves(bonzaURL);
    });

    cy.get('[data-testid="inventory-paste-action"]').click();
    cy.wait("@bonzaProductIngest");
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-create-paste-success"]').should(
      "contain",
      "Provider data loaded"
    );
    cy.get('[data-testid="inventory-item-title"]').should(
      "have.value",
      "BONZA MUG WHITE"
    );
    cy.get('[data-testid="inventory-item-brand"]').should("have.value", "AFX");
    cy.get('[data-testid="inventory-item-description"]')
      .should("contain.value", "AUD 9.95")
      .and("contain.value", "BONZA-MUG.jpg");
    cy.get('[data-testid="inventory-item-create-submit"]').click();
    cy.wait("@createBonzaItem");
    cy.wait("@itemsBonzaPaste");
    cy.get('[data-testid="inventory-item-row-item-bonza-mug"]').should("exist");
  });

  it("UI-SCREEN-INVENTORY-BONZA-002 preserves unsupported pasted URL for manual creation", () => {
    const unsupportedURL = "https://example.com/not-a-provider/item-1";
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsUnsupportedPaste");
    cy.intercept("POST", "/api/providers/product-url/ingest", (req) => {
      expect(req.body).to.deep.eq({
        url: unsupportedURL,
        capture_for_review: true,
      });
      req.reply({
        statusCode: 200,
        body: {
        mode: "provider_product_url_ingest",
        error: "unsupported_provider_url",
      },
      });
    }).as("unsupportedProviderIngest");

    signIn();
    cy.wait("@itemsUnsupportedPaste");
    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-create-paste-input"]').type(unsupportedURL);
    cy.get('[data-testid="inventory-create-paste-process"]').click();
    cy.wait("@unsupportedProviderIngest");
    cy.get('[data-testid="inventory-create-paste-error"]').should(
      "contain",
      "not supported"
    );
    cy.get('[data-testid="inventory-create-paste-input"]').should(
      "have.value",
      unsupportedURL
    );
  });

  it("UI-SCREEN-INVENTORY-BONZA-004 preserves failed Bonza ingest input and manual source URL context", () => {
    const bonzaURL = "https://bonzaslotcars.com.au/product/bonza-mug-white/";
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsFailedBonzaPaste");
    cy.intercept("POST", "/api/providers/product-url/ingest", (req) => {
      expect(req.body).to.deep.eq({ url: bonzaURL, capture_for_review: true });
      req.reply({
        statusCode: 502,
        body: {
        mode: "provider_product_url_ingest",
        error: "failed_to_ingest_bonza_product_url",
        guidance:
          "Static product extraction was attempted first but the storefront did not return usable public product data. Open Bonza yourself in the paired Browser Companion and sync the rendered product; Cabinet does not run a hidden browser or export session data.",
        review_capture_persisted: true,
        review_capture: {
          item_id: "item-bonza-manual-review",
          title: "Manual review: bonzaslotcars.com.au/product/bonza-mug-white",
          status: "manual_review",
          source_url: bonzaURL,
          fallback_state: "browser_companion_user_present",
        },
      },
      });
    }).as("failedBonzaProductIngest");

    signIn();
    cy.wait("@itemsFailedBonzaPaste");
    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-create-paste-input"]').type(bonzaURL);
    cy.get('[data-testid="inventory-create-paste-process"]').click();
    cy.wait("@failedBonzaProductIngest");
    cy.get('[data-testid="inventory-create-paste-error"]')
      .should("have.attr", "role", "alert")
      .and("contain", "Static product extraction was attempted first")
      .and(
        "contain",
        "Captured for review as Manual review: bonzaslotcars.com.au/product/bonza-mug-white"
      );
    cy.get('[data-testid="inventory-create-paste-input"]').should(
      "have.value",
      bonzaURL
    );
    cy.get('[data-testid="inventory-item-description"]').should(
      "contain.value",
      bonzaURL
    );
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
  });

  it("UI-SCREEN-INVENTORY-BONZA-003 shows duplicate candidates and can open the existing item", () => {
    const bonzaURL = "https://bonzaslotcars.com.au/product/bonza-mug-white/";
    const existingItem = {
      id: "item-existing-bonza-mug",
      part_number: "BONZA-19603",
      title: "Existing Bonza Mug",
      status: "active",
      condition: "",
      category: "AFX ACCESSORIES HO, MERCHANDISE",
      item_type: "General",
      packaging_grade_type: "",
      brand: "AFX",
      priority: "medium",
      description: "",
      notes: "",
      tags: ["provider-ingest"],
      source_urls: [bonzaURL],
    };

    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [existingItem] },
    }).as("itemsWithDuplicate");
    cy.intercept("POST", "/api/providers/product-url/ingest", (req) => {
      expect(req.body).to.deep.eq({ url: bonzaURL, capture_for_review: true });
      req.reply({
        statusCode: 200,
        body: {
        mode: "provider_product_url_ingest",
        provider: "bonzaslotcars",
        family: "woocommerce",
        draft: {
          provider_product_id: "19603",
          title: "BONZA MUG WHITE",
          source_url: bonzaURL,
          categories: ["AFX ACCESSORIES HO", "MERCHANDISE"],
          attributes: { Brand: "AFX" },
          image_urls: [],
        },
        evidence: {
          provider_product_id: "19603",
          original_url: bonzaURL,
          normalized_url: bonzaURL,
        },
        duplicates: [
          {
            item_id: existingItem.id,
            title: existingItem.title,
            source_urls: existingItem.source_urls,
            reasons: ["source_url", "provider_product_id"],
          },
        ],
      },
      });
    }).as("duplicateProviderIngest");

    signIn();
    cy.wait("@itemsWithDuplicate");
    cy.get('[data-testid="inventory-new-action"]').click();
    cy.get('[data-testid="inventory-create-paste-input"]').type(bonzaURL);
    cy.get('[data-testid="inventory-create-paste-process"]').click();
    cy.wait("@duplicateProviderIngest");
    cy.get('[data-testid="inventory-create-duplicate-warning"]')
      .should("be.visible")
      .should("contain", "Existing Bonza Mug")
      .and("contain", "source_url");
    cy.get('[data-testid="inventory-item-create-submit"]').should("exist");
    cy.get('[data-testid="inventory-create-open-duplicate"]')
      .scrollIntoView()
      .click();
    cy.get('[data-testid="inventory-item-edit-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-item-title"]').should(
      "have.value",
      "Existing Bonza Mug"
    );
  });
});
