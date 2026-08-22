describe("ui-foundation-theme-rtl-i18n", () => {
  function enterHomeWithLocalSession(options: { initialLanguage?: string } = {}) {
    cy.viewport(1512, 967);
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.e2eEnsureSignedOut();
      cy.stubLocalServerSession(profile_id);
      cy.setCookie("sidebar_state", "true");
      cy.request("PUT", "/api/profiles/active", { profile_id }).its("status").should("eq", 200);
      cy.visit("/sign-in?redirect=%2Fdashboard", {
        onBeforeLoad(win) {
          win.localStorage.setItem(`cabinet.workspace.${profile_id}`, "1");
          if (options.initialLanguage) {
            win.localStorage.setItem("i18nextLng", options.initialLanguage);
          }
        },
      });
      cy.contains("button", "Open local workspace").click();
      cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard");
      cy.get("body").then(($body) => {
        const profileButton = `Use ${profile_name}`;
        if ($body.text().includes(profileButton)) {
          cy.contains("button", profileButton).click();
        }
      });
    });
  }

  function switchLanguage(code: "en" | "zh" | "ja") {
    cy.get('[data-testid="header-language-switch-trigger"]').filter(":visible").click();
    cy.get(`[data-testid="header-language-option-${code}"]`).filter(":visible").click();
  }

  it("UI-FOUNDATION-THEME-RTL-I18N-001 persists selected theme across reload", () => {
    enterHomeWithLocalSession();

    cy.contains("button", "Toggle theme").click();
    cy.contains('[role="menuitem"]', "Dark").click();
    cy.get("html").should("have.class", "dark");

    cy.reload();
    cy.get("html").should("have.class", "dark");
  });

  it("UI-FOUNDATION-THEME-RTL-I18N-004 provides header locale switch and fallback locale behavior", () => {
    enterHomeWithLocalSession({ initialLanguage: "zz" });

    cy.get('[data-testid="header-language-switch-trigger"]')
      .should("be.visible")
      .and("contain", "EN")
      .click();
    cy.get('[data-testid="header-language-option-en"]').should("be.visible").click();
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should("contain", "Home");
    cy.contains('h1', 'Home').should('be.visible');
  });

  it("UI-FOUNDATION-THEME-RTL-I18N-002 keeps shell labels safe when active locale lacks specific keys", () => {
    enterHomeWithLocalSession();
    switchLanguage("ja");
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should("contain", "ホーム");
    cy.get('[data-testid="sidebar-nav-link-inventory"]').should("contain", "在庫");
  });

  it("UI-FOUNDATION-THEME-RTL-I18N-003 keeps layout direction stable for supported non-RTL locales", () => {
    enterHomeWithLocalSession();
    switchLanguage("ja");
    cy.get("html").should("have.attr", "dir", "ltr");
  });
});
