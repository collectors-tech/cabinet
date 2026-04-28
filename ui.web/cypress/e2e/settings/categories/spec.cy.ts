describe("settings/categories", () => {
  function signInToCategories() {
    cy.visit("/sign-in?redirect=%2Fsettings%2Fcategories");
    cy.get('input[name="email"]').clear().type("e2e-settings@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/settings\/categories\/?$/
    );
  }

  beforeEach(() => {
    cy.clearCookies();
    cy.clearLocalStorage();
  });

  it("manages the reusable inventory category list for the active profile", () => {
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "profile-categories" },
    }).as("activeProfile");

    cy.intercept("GET", "/api/profiles/profile-categories/settings", {
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

    cy.intercept("PUT", "/api/profiles/profile-categories/settings", (req) => {
      const settings = req.body?.settings ?? {};
      const categories = JSON.parse(
        settings["inventory.category-options.v1"] ?? "[]"
      );
      expect(categories).to.deep.equal(["General", "Model Kit", "Garage Kit"]);
      req.reply({ statusCode: 200, body: { settings } });
    }).as("saveProfileSettings");

    signInToCategories();
    cy.wait("@activeProfile");
    cy.wait("@profileSettings");

    cy.contains("h3", "Categories").should("be.visible");
    cy.get('[data-testid="settings-categories-list"]').should("contain", "Cars");
    cy.get('[data-testid="settings-categories-new"]').type("Garage Kit");
    cy.get('[data-testid="settings-categories-add"]').click();
    cy.get('[data-testid="settings-category-remove-Cars"]').click();
    cy.get('[data-testid="settings-categories-save"]').click();
    cy.wait("@saveProfileSettings");
    cy.get('[data-testid="settings-categories-status"]').should(
      "contain",
      "Saved categories."
    );
  });
});
