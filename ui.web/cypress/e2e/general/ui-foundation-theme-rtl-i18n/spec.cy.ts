describe("ui-foundation-theme-rtl-i18n", () => {
  function signInToHome() {
    cy.clearCookies();
    cy.clearLocalStorage();
    cy.visit("/sign-in?redirect=%2F");
    cy.get('input[name="email"]').clear().type("e2e-theme-i18n@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should("eq", "/");
  }

  function switchLanguage(code: "en" | "ar") {
    cy.get('[data-testid="header-language-switch-trigger"]').click();
    cy.get(`[data-testid="header-language-option-${code}"]`).click();
  }

  it("UI-FOUNDATION-THEME-RTL-I18N-001 persists selected theme across reload", () => {
    signInToHome();

    cy.contains("button", "Toggle theme").click();
    cy.contains('[role="menuitem"]', "Dark").click();
    cy.get("html").should("have.class", "dark");

    cy.reload();
    cy.get("html").should("have.class", "dark");
  });

  it("UI-FOUNDATION-THEME-RTL-I18N-004 provides header locale switch and fallback locale behavior", () => {
    cy.clearCookies();
    cy.clearLocalStorage();
    cy.visit("/sign-in?redirect=%2F", {
      onBeforeLoad(win) {
        win.localStorage.setItem("i18nextLng", "zz");
      },
    });
    cy.get('input[name="email"]').clear().type("e2e-theme-i18n@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should("eq", "/");

    cy.get('[data-testid="header-language-switch-trigger"]')
      .should("be.visible")
      .and("contain", "EN")
      .click();
    cy.get('[data-testid="header-language-option-en"]').should("be.visible").click();
    cy.contains("button", "Dashboard").should("be.visible");
  });

  it("UI-FOUNDATION-THEME-RTL-I18N-002 keeps shell labels safe when active locale lacks specific keys", () => {
    signInToHome();
    switchLanguage("ar");
    cy.contains("button", "Dashboard").should("be.visible");
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should("be.visible");
  });

  it("UI-FOUNDATION-THEME-RTL-I18N-003 mirrors layout direction when RTL locale is selected", () => {
    signInToHome();
    switchLanguage("ar");
    cy.get("html").should("have.attr", "dir", "rtl");
  });
});
