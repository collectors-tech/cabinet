type E2EBootstrapOptions = {
  minimalProfile?: boolean;
};

type E2EBootstrapState = {
  profile_id: string;
  profile_name: string;
  item_ids: string[];
  query_set_id: string;
  thread_id: string;
};

type E2ESetupState = "missing" | "present";

declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      e2eReset(): Chainable<void>;
      e2eBootstrap(options?: E2EBootstrapOptions): Chainable<E2EBootstrapState>;
      e2eEnsureSignedOut(): Chainable<void>;
      e2eSetSetupState(state: E2ESetupState): Chainable<void>;
      e2eCompleteSetupHelper(overrides?: {
        instance_name?: string;
        profile_key?: string;
        runtime?: { mode?: "auto" | "fixed"; fixed_port?: number };
      }): Chainable<void>;
      useBootstrappedProfile(
        profileId: string,
        profileName: string,
        options?: {
          workspace?: boolean;
          path?: string;
          shellWorkspace?: "navigation" | "assistant" | "inbox";
        }
      ): Chainable<void>;
    }
  }
}

const AUTH_COOKIE_NAMES = ["thisisjustarandomstring", "cabinet_auth_user"] as const;

function expireAuthCookies(win: Window) {
  AUTH_COOKIE_NAMES.forEach((name) => {
    win.document.cookie = `${name}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
  });
}

Cypress.Commands.add("e2eReset", () => {
  cy.request("POST", "/api/test/reset", {}).its("status").should("eq", 200);
});

Cypress.Commands.add("e2eEnsureSignedOut", () => {
  cy.visit("/sign-in", {
    onBeforeLoad(win) {
      win.localStorage.clear();
      win.sessionStorage.clear();
      expireAuthCookies(win);
    },
  });

  cy.clearCookies();
  cy.clearLocalStorage();
  cy.window({ log: false }).then((win) => {
    win.sessionStorage.clear();
    expireAuthCookies(win);
  });
  cy.getCookie("thisisjustarandomstring").should("not.exist");
  cy.getCookie("cabinet_auth_user").should("not.exist");
  cy.location("pathname", { timeout: 15000 }).should("eq", "/sign-in");
});

Cypress.Commands.add("e2eBootstrap", (options: E2EBootstrapOptions = {}) => {
  return cy
    .request<E2EBootstrapState>("POST", "/api/test/bootstrap", {
      minimal_profile: Boolean(options.minimalProfile),
    })
    .then((resp) => {
      expect(resp.status).to.eq(200);
      expect(resp.body.profile_id).to.eq("e2e-profile-001");
      expect(resp.body.profile_name).to.eq("E2E Local");
      return resp.body;
    });
});

Cypress.Commands.add("e2eSetSetupState", (state: E2ESetupState) => {
  cy.request("POST", "/api/test/runtime/setup-status", { state }).its("status").should("eq", 200);
});

Cypress.Commands.add("e2eCompleteSetupHelper", (overrides = {}) => {
  const instanceName = overrides.instance_name ?? "E2E Setup Helper";
  const profileKey = overrides.profile_key ?? "e2e-setup-helper";
  const runtimeMode = overrides.runtime?.mode ?? "auto";
  const fixedPort = overrides.runtime?.fixed_port;
  return cy
    .request("POST", "/api/runtime/setup-complete", {
      instance_name: instanceName,
      profile_key: profileKey,
      auth: { mode: "local", clerk_publishable_key: "", clerk_sign_in_url: "" },
      storage: { mode: "exe_local", data_dir: "", media_dir: "", portable_mode: false },
      runtime: {
        mode: runtimeMode,
        fixed_port: runtimeMode === "fixed" ? fixedPort ?? 17880 : 0,
      },
      features: { scanner: true, providers: true, chat: true },
    })
    .then((resp) => {
      expect(resp.status).to.eq(200);
      expect(resp.body.setup_required).to.eq(false);
    });
});

Cypress.Commands.add("useBootstrappedProfile", (profileId: string, profileName: string, options?: {
  workspace?: boolean;
  path?: string;
  shellWorkspace?: "navigation" | "assistant" | "inbox";
}) => {
  const withWorkspace = options?.workspace ?? true;
  const targetPath = options?.path ?? "/";
  const normalizedTarget = targetPath.endsWith("/") && targetPath.length > 1 ? targetPath.slice(0, -1) : targetPath;
  const escapedTarget = normalizedTarget.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const targetPathRegex = normalizedTarget === "/" ? /^\/$/ : new RegExp(`^${escapedTarget}\\/?$`);
  cy.request("PUT", "/api/profiles/active", { profile_id: profileId }).its("status").should("eq", 200);
  cy.visit(`/sign-in?redirect=${encodeURIComponent(targetPath)}`, {
    onBeforeLoad(win) {
      if (withWorkspace) {
        win.localStorage.setItem(`cabinet.workspace.${profileId}`, "1");
      }
      if (options?.shellWorkspace) {
        win.localStorage.setItem(`cabinet.shell.workspace.active.${profileId}`, options.shellWorkspace);
      }
    },
  });
  cy.get('input[name="email"]').clear().type("e2e-login-session@example.com");
  cy.get('input[name="password"]').clear().type("password123");
  cy.contains("button", "Sign in").click();
  cy.location("pathname", { timeout: 15000 }).should("match", targetPathRegex);
  cy.get("body").then(($body) => {
    const preferredLabel = `Use ${profileName}`;
    if ($body.text().includes(preferredLabel)) {
      cy.contains("button", preferredLabel).click();
      return;
    }

    const buttonWithUsePrefix = $body
      .find("button")
      .toArray()
      .map((node) => (node.textContent ?? "").replace(/\s+/g, " ").trim())
      .find((text) => /^Use\s+/i.test(text));

    if (buttonWithUsePrefix) {
      cy.contains("button", buttonWithUsePrefix).click();
    }
  });
});

export {};
