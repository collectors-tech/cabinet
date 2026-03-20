describe("ui-login-session", () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.request("POST", "/api/test/runtime/setup-status", { state: "present" })
      .its("status")
      .should("eq", 200);
  });

  it("UI-LOGIN-SESSION-001 redirects unauthenticated access to sign-in and returns to target after login", () => {
    cy.clearCookies();
    cy.clearLocalStorage();

    cy.visit("/inventory/");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%2F");

    cy.get('input[name="email"]').type("e2e-login-session@example.com");
    cy.get('input[name="password"]').type("password123");
    cy.contains("button", "Sign in").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/inventory\/?$/
    );
  });

  it("UI-LOGIN-SESSION-002 keeps inline validation errors and allows retry without refresh", () => {
    cy.clearCookies();
    cy.clearLocalStorage();

    cy.visit("/sign-in");
    cy.get('input[name="email"]').type("invalid-email");
    cy.get('input[name="password"]').type("short");
    cy.contains("button", "Sign in").click();

    cy.location("pathname").should("eq", "/sign-in");
    cy.contains(/invalid email|please enter your email/i).should("be.visible");
    cy.contains("Password must be at least 7 characters long").should(
      "be.visible"
    );

    cy.get('input[name="email"]').clear().type("e2e-login-session@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^(\/|\/inventory\/?|\/_authenticated\/?)$/
    );
  });

  it("UI-LOGIN-SESSION-003 switches active profile after login and uses selected profile scope for subsequent API calls", () => {
    cy.clearCookies();
    cy.clearLocalStorage();

    let activeProfile = { id: "p1", name: "Default" };

    cy.intercept("GET", "/api/profiles", {
      statusCode: 200,
      body: {
        profiles: [
          { id: "p1", name: "Default" },
          { id: "p2", name: "Perf S2" },
        ],
      },
    }).as("profiles");

    cy.intercept("GET", "/api/profiles/active", (req) => {
      req.reply({ statusCode: 200, body: activeProfile });
    }).as("activeProfile");

    cy.intercept("PUT", "/api/profiles/active", (req) => {
      expect(req.body.profile_id).to.eq("p2");
      activeProfile = { id: "p2", name: "Perf S2" };
      req.reply({ statusCode: 200, body: activeProfile });
    }).as("setActiveProfile");

    cy.intercept("GET", "/api/providers/registry", {
      statusCode: 200,
      body: { providers: [] },
    }).as("registry");

    cy.intercept("GET", "/api/profiles/p2/settings", {
      statusCode: 200,
      body: { settings: {} },
    }).as("profile2Settings");

    cy.visit("/sign-in?redirect=%2F");
    cy.get('input[name="email"]').type("e2e-login-session@example.com");
    cy.get('input[name="password"]').type("password123");
    cy.contains("button", "Sign in").click();

    cy.location("pathname", { timeout: 15000 }).should("eq", "/");
    cy.wait("@profiles");

    cy.get('[data-testid="team-switcher-trigger"]').click();
    cy.get('[data-testid="team-option-perf-s2"]').click();
    cy.wait("@setActiveProfile");
    cy.get('[data-testid="active-profile-name"]').should("contain", "Perf S2");

    cy.contains("a", "Integrations").click();
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/integrations\/?$/);
    cy.wait("@registry");
    cy.wait("@profile2Settings");
  });

  it("UI-LOGIN-SESSION-004 provisions active profile on first-run and avoids active_profile_404 across core routes", () => {
    cy.visit("/sign-in?redirect=%2Fsettings%2Fdisplay");
    cy.get('input[name="email"]').type("e2e-first-run-profile@example.com");
    cy.get('input[name="password"]').type("password123");
    cy.contains("button", "Sign in").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/settings\/display\/?$/
    );
    cy.contains(/active_profile_404|active_profile_not_set/i).should("not.exist");

    ["/chats", "/integrations", "/reports", "/users"].forEach((path) => {
      cy.visit(path);
      cy.location("pathname", { timeout: 15000 }).should(
        "match",
        new RegExp(`^${path}\\/?$`)
      );
      cy.contains(/active_profile_404|active_profile_not_set/i).should("not.exist");
    });
  });

  it("UI-LOGIN-SESSION-005 redirects base unauthenticated entry to clean sign-in while preserving deep-link redirects", () => {
    cy.clearCookies();
    cy.clearLocalStorage();

    cy.visit("/");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("eq", "");

    cy.get('input[name="email"]').type("e2e-clean-root-entry@example.com");
    cy.get('input[name="password"]').type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/dashboard\/?$/);

    cy.clearCookies();
    cy.clearLocalStorage();

    cy.visit("/inventory/");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%2F");
  });
});
