describe("ui-login-session", () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.request("POST", "/api/test/runtime/setup-status", { state: "present" })
      .its("status")
      .should("eq", 200);
    cy.e2eEnsureSignedOut();
  });

  it("UI-LOGIN-SESSION-001 redirects unauthenticated access to sign-in and returns to target after login", () => {
    cy.visit("/inventory/");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%2F");

    cy.contains("button", "Open local workspace").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/inventory\/?$/
    );
    cy.getCookie("thisisjustarandomstring")
      .its("value")
      .should("not.contain", "mock-access-token")
      .and("not.contain", "mock-passkey-access-token");
  });

  it("UI-LOGIN-SESSION-002 keeps local-device auth boundary truthful and avoids simulated password success", () => {
    cy.visit("/sign-in");

    cy.location("pathname").should("eq", "/sign-in");
    cy.get('[data-testid="local-device-auth-boundary"]').should("be.visible");
    cy.contains(/does not verify a password/i).should("be.visible");
    cy.get('input[name="password"]').should("not.exist");
    cy.contains("button", "Sign in").should("not.exist");

    cy.contains("button", "Open local workspace").click();

    cy.location("pathname", { timeout: 15000 }).should("match", /^\/dashboard\/?$/);
    cy.window().then((win) => {
      const history = JSON.parse(
        win.localStorage.getItem("cabinet.toastHistory.v1") ?? "[]"
      ) as Array<Record<string, unknown>>;
      const signInRecord = history.find((record) => record.id === "auth-sign-in");

      expect(signInRecord).to.include({
        level: "success",
        title: "Local workspace opened",
        source_label: "Auth sign-in",
        category: "auth",
      });
    });
  });

  it("UI-LOGIN-SESSION-011 presents the Cabinet ZITADEL boundary without password fields", () => {
    cy.request("POST", "/api/test/auth/provider-options", {
      identity_mode: "zitadel",
      zitadel_configured: true,
      zitadel_login_path: "/api/auth/zitadel/login",
      providers: [],
    })
      .its("status")
      .should("eq", 200);

    cy.visit("/sign-in?redirect=%2Fsettings%2Fdisplay");

    cy.get('[data-testid="zitadel-auth-boundary"]').should("be.visible");
    cy.contains("Cabinet secure account").should("be.visible");
    cy.contains("button", "Continue securely").should("be.enabled");
    cy.get('input[name="email"]').should("not.exist");
    cy.get('input[name="password"]').should("not.exist");
    cy.contains(/opaque secure session cookie/i).should("be.visible");
  });

  it("UI-LOGIN-SESSION-003 switches active profile after login and uses selected profile scope for subsequent API calls", () => {
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

    cy.visit("/sign-in?redirect=%2Fdashboard");
    cy.contains("button", "Open local workspace").click();

    cy.location("pathname", { timeout: 15000 }).should("match", /^\/dashboard\/?$/);
    cy.wait("@profiles");

    cy.get('[data-testid="team-switcher-trigger"]').click();
    cy.get('[data-testid="team-option-perf-s2"]').click();
    cy.wait("@setActiveProfile");
    cy.get('[data-testid="active-profile-name"]').should("contain", "Perf S2");

    cy.get('[data-testid="sidebar-nav-link-integrations"]')
      .scrollIntoView()
      .should("be.visible")
      .and("have.attr", "aria-label", "Integrations")
      .click();
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/integrations\/?$/);
    cy.wait("@registry");
    cy.wait("@profile2Settings");
  });

  it("UI-LOGIN-SESSION-004 provisions active profile on first-run and avoids active_profile_404 across core routes", () => {
    cy.visit("/sign-in?redirect=%2Fsettings%2Fdisplay");
    cy.contains("button", "Open local workspace").click();

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
    cy.visit("/");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("eq", "");

    cy.contains("button", "Open local workspace").click();
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/dashboard\/?$/);

    cy.e2eEnsureSignedOut();

    cy.visit("/inventory/");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%2F");
  });

  it("UI-LOGIN-SESSION-006 clears session state on direct sign-out route and re-gates protected routes", () => {
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });

    cy.visit("/sign-out?redirect=%2Finventory%2F");
    cy.location("pathname", { timeout: 15000 }).should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%2F");
    cy.getCookie("thisisjustarandomstring").should("not.exist");

    cy.visit("/inventory/");
    cy.location("pathname", { timeout: 15000 }).should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%2F");
  });

  it("UI-LOGIN-SESSION-007 re-gates the dashboard after sign-out", () => {
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/dashboard" });
    });

    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard");
    cy.contains(/dashboard/i).should("be.visible");

    cy.visit("/sign-out?redirect=%2Fdashboard");
    cy.location("pathname", { timeout: 15000 }).should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=%2Fdashboard");
    cy.getCookie("thisisjustarandomstring").should("not.exist");

    cy.visit("/dashboard");
    cy.location("pathname", { timeout: 15000 }).should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=%2Fdashboard");
    cy.contains(/dashboard/i).should("not.exist");
  });

  it("UI-LOGIN-SESSION-009 recovers cookie-backed session after protected-route reload", () => {
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/inventory\/?$/
    );
    cy.getCookie("thisisjustarandomstring").should("exist");

    cy.reload();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/inventory\/?$/
    );
    cy.contains("button", "Open local workspace").should("not.exist");
    cy.contains("button", "Sign in").should("not.exist");
    cy.contains(/inventory/i).should("be.visible");
  });

  it("UI-LOGIN-SESSION-010 preserves protected-route query state through sign-in redirect", () => {
    cy.visit("/inventory?view=table");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%3Fview%3Dtable");

    cy.contains("button", "Open local workspace").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/inventory\/?$/
    );
    cy.location("search").should("eq", "?view=table");
  });

  it("UI-LOGIN-SESSION-008 keeps sign-in copy focused while preserving account and legal links", () => {
    cy.visit("/sign-in");

    cy.contains("Sign in to unlock your Cabinet workspace.").should("not.exist");
    cy.get('[data-testid="sign-in-profile-guidance"]').should("not.exist");
    cy.get('[data-testid="local-device-auth-boundary"]').should("be.visible");
    cy.get('input[name="email"]').should("not.exist");
    cy.get('input[name="password"]').should("not.exist");
    cy.contains("button", "Open local workspace").should("be.visible");
    cy.contains("a", "Create account").should("have.attr", "href", "/sign-up");
    cy.contains("a", "Forgot password?").should(
      "have.attr",
      "href",
      "/forgot-password"
    );
    cy.contains("a", "Terms of Service").should("have.attr", "href", "/terms");
    cy.contains("a", "Privacy Policy").should("have.attr", "href", "/privacy");
  });
});
