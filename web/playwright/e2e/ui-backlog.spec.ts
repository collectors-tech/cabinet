import { expect, test } from "@playwright/test";

type WizardMockOptions = {
  registerBeginFailsOnce?: boolean;
  sampleFailsOnce?: boolean;
  requiresRegistration?: boolean;
};

async function installWizardMocks(page: Parameters<typeof test>[0]["page"], options: WizardMockOptions = {}) {
  const state = {
    items: [{ id: "i1", part_number: "PN-001", title: "Existing Item", brand: "AFX" }] as Array<{ id: string; part_number: string; title: string; brand?: string }>,
    sampleCalls: 0,
    registerBeginCalls: 0,
    registerBeginFailsOnce: Boolean(options.registerBeginFailsOnce),
    sampleFailsOnce: Boolean(options.sampleFailsOnce),
    requiresRegistration: options.requiresRegistration ?? true,
  };

  await page.route("**/api/**", async (route) => {
    const req = route.request();
    const url = new URL(req.url());
    const path = url.pathname;
    const method = req.method();
    const body = req.postDataJSON?.() as Record<string, unknown> | undefined;
    const respondJSON = (payload: unknown, status = 200) =>
      route.fulfill({ status, contentType: "application/json", body: JSON.stringify(payload) });

    if (path === "/api/profiles" && method === "GET") {
      return respondJSON({ profiles: [{ id: "p1", name: "Default" }] });
    }
    if (path === "/api/profiles/active" && method === "PUT") {
      return respondJSON({ id: "p1", name: "Default" });
    }
    if (path === "/api/profiles/p1/storage" && method === "GET") {
      return respondJSON({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" });
    }
    if (path === "/api/auth/requirements" && method === "GET") {
      return respondJSON({ requires_registration: state.requiresRegistration });
    }
    if (path === "/api/items" && method === "GET") {
      return respondJSON({ items: state.items });
    }
    if (path === "/api/items" && method === "POST") {
      const next = {
        id: `i${state.items.length + 1}`,
        part_number: String(body?.part_number ?? ""),
        title: String(body?.title ?? ""),
        brand: String(body?.brand ?? ""),
      };
      state.items.push(next);
      return respondJSON(next, 201);
    }
    if (path === "/api/dashboard" && method === "GET") {
      return respondJSON({ new_discoveries: 2, wishlist_hits: 1, price_drops: 1, recently_added: 3, total_items: 10, total_instances: 14 });
    }
    if (path === "/api/scanner/run/scheduled" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/scanner/failures" && method === "GET") {
      return respondJSON({ failures: [{ id: "f1", query_set_id: "q1", reason: "rate_limited", attempts: 2 }] });
    }
    if (path === "/api/provider/health" && method === "GET") {
      return respondJSON({ provider: "ebay", state: "healthy", healthy: true });
    }
    if (path === "/api/matching/run" && method === "POST") {
      return respondJSON({ processed: 3 });
    }
    if (path === "/api/matching/results" && method === "GET") {
      return respondJSON({ results: [{ candidate_id: "c1", state: "matched" }] });
    }
    if (path === "/api/pricing/track" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/pricing/history" && method === "GET") {
      return respondJSON({ history: [{ day: "2026-02-21", min: 10, median: 11, latest: 12 }] });
    }
    if (path === "/api/pricing/stats" && method === "GET") {
      return respondJSON({ min: 10, median: 11, latest: 12 });
    }
    if (path === "/api/pricing/trend" && method === "GET") {
      return respondJSON({ trend: "down" });
    }
    if (path === "/api/pricing/snapshot/run" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/wishlist/hits" && method === "GET") {
      return respondJSON({ hits: [{ item_id: "i1", listing_id: "l1", title: "Hit Item", price: 18 }] });
    }
    if (path === "/api/backup/list" && method === "GET") {
      return respondJSON({ backups: ["/tmp/backups/cabinet-backup-20260223-120000.db"] });
    }
    if (path === "/api/backup/restore" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/profiles/p1/license" && method === "GET") {
      return respondJSON({ license_json: "{\"tier\":\"pro\"}" });
    }
    if (path === "/api/profiles/p1/license" && method === "PUT") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/license/import" && method === "POST") {
      const payload = (body?.license ?? {}) as Record<string, unknown>;
      if (String(payload.payload_base64 || "") === "bad") {
        return respondJSON({ error: "failed_to_import_license" }, 400);
      }
      return respondJSON({ ok: true });
    }
    if (path === "/api/license/status" && method === "GET") {
      return respondJSON({ state: "valid", tier: "pro", features: ["ai_assist", "price_tracking"], expires_at: "2030-01-01T00:00:00Z" });
    }
    if (path === "/api/logs/debug" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/auth/webauthn/register/begin" && method === "POST") {
      state.registerBeginCalls += 1;
      if (state.registerBeginFailsOnce && state.registerBeginCalls === 1) {
        return respondJSON({ error: "failed" }, 500);
      }
      return respondJSON({ session_id: "sess-reg-1", options: {} });
    }
    if (path === "/api/auth/webauthn/register/finish" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/auth/webauthn/login/begin" && method === "POST") {
      return respondJSON({ session_id: "sess-login-1", options: {} });
    }
    if (path === "/api/auth/webauthn/login/finish" && method === "POST") {
      return respondJSON({ session_token: "token-1" });
    }
    if (path === "/api/onboarding/sample-data" && method === "POST") {
      state.sampleCalls += 1;
      if (state.sampleFailsOnce && state.sampleCalls === 1) {
        return respondJSON({ error: "failed" }, 500);
      }
      const created = state.sampleCalls === 1 ? 2 : 0;
      return respondJSON({ created_items: created });
    }
    if (path === "/api/profiles/p1/settings" && method === "PUT") {
      return respondJSON({ settings: (body?.settings as Record<string, unknown>) ?? {} });
    }
    return respondJSON({});
  });

  return state;
}

async function openStarterWizard(page: Parameters<typeof test>[0]["page"]) {
  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "0");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();
  await expect(page.getByRole("heading", { name: /starter onboarding wizard/i })).toBeVisible();
}

test("api health smoke (real backend)", async ({ request }) => {
  const health = await request.get("/healthz");
  expect(health.ok()).toBeTruthy();
});

test("left navigation switches visible advanced-workspace screens (desktop + mobile)", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: false });
  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "1");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();

  await expect(page.getByRole("heading", { name: /^dashboard$/i })).toBeVisible();
  await page.getByRole("button", { name: /^dashboard$/i }).first().click();
  await expect(page.getByRole("button", { name: /^dashboard$/i }).first()).toHaveAttribute("aria-current", "page");
  await expect(page.getByTestId("screen-dashboard")).toBeVisible();
  await expect(page.getByText(/new discoveries: 2/i)).toBeVisible();

  await page.getByRole("button", { name: /^collection$/i }).first().click();
  await expect(page.getByRole("heading", { name: /^collection$/i })).toBeVisible();
  await expect(page.getByRole("button", { name: /^collection$/i }).first()).toHaveAttribute("aria-current", "page");
  await expect(page.getByTestId("screen-collection")).toBeVisible();

  await page.setViewportSize({ width: 800, height: 900 });
  await page.getByRole("button", { name: /open navigation menu/i }).click();
  const drawer = page.getByRole("dialog", { name: /navigation menu/i });
  await expect(drawer).toBeVisible();
  await drawer.getByRole("button", { name: /^scanner$/i }).click();
  await expect(page.getByRole("dialog", { name: /navigation menu/i })).toHaveCount(0);
  await expect(page.getByRole("heading", { name: /discovery scanner/i })).toBeVisible();
  await expect(page.getByTestId("screen-scanner")).toBeVisible();
  await page.getByRole("button", { name: /run scheduled/i }).click();
  await expect(page.getByText(/scheduled run: scheduled_scans_triggered/i)).toBeVisible();
  await page.getByRole("button", { name: /load scanner failures/i }).click();
  await expect(page.getByText(/failure: q1 \/ rate_limited \/ attempts 2/i)).toBeVisible();
  await page.getByRole("button", { name: /check provider health/i }).click();
  await expect(page.getByText(/provider health: ebay \/ healthy/i)).toBeVisible();
  await page.getByRole("button", { name: /run matching/i }).click();
  await expect(page.getByText(/matching run status: matching_run_ok:3/i)).toBeVisible();
  await expect(page.getByText(/matched: 1/i)).toBeVisible();

  await page.setViewportSize({ width: 1200, height: 900 });
  await page.getByRole("button", { name: /^pricing$/i }).first().click();
  await expect(page.getByTestId("screen-pricing")).toBeVisible();
  await page.getByRole("button", { name: /track pricing/i }).click();
  await expect(page.getByText(/pricing track status: pricing_track_enabled/i)).toBeVisible();
  await page.getByRole("button", { name: /load pricing history/i }).click();
  await expect(page.getByText(/pricing history points: 1/i)).toBeVisible();
  await page.getByRole("button", { name: /load pricing stats/i }).click();
  await expect(page.getByText(/pricing stats loaded: yes/i)).toBeVisible();
  await page.getByRole("button", { name: /load pricing trend/i }).click();
  await expect(page.getByText(/pricing trend loaded: yes/i)).toBeVisible();
  await page.getByRole("button", { name: /run pricing snapshot/i }).click();
  await expect(page.getByText(/snapshot status: pricing_snapshot_completed/i)).toBeVisible();
  await page.getByRole("button", { name: /load wishlist hits/i }).click();
  await expect(page.getByText(/wishlist hit: i1 \/ hit item \/ 18/i)).toBeVisible();

  await page.getByRole("button", { name: /^settings$/i }).first().click();
  await expect(page.getByTestId("screen-settings")).toBeVisible();
  await page.getByRole("button", { name: /load backups/i }).click();
  await expect(page.getByText(/backup count: 1/i)).toBeVisible();
  await page.getByRole("button", { name: /restore selected backup/i }).click();
  await expect(page.getByText(/admin error: restore_confirmation_required/i)).toBeVisible();
  await page.getByLabel(/confirm restore/i).check();
  await page.getByRole("button", { name: /restore selected backup/i }).click();
  await expect(page.getByText(/settings status: backup_restored/i)).toBeVisible();

  await page.getByRole("button", { name: /load profile license/i }).click();
  await expect(page.getByLabel(/profile license json/i)).toHaveValue(/"tier":"pro"/i);
  await page.getByLabel(/license payload base64/i).fill("good");
  await page.getByLabel(/license signature base64/i).fill("sig");
  await page.getByRole("button", { name: /import license file/i }).click();
  await expect(page.getByText(/license import status: license_imported/i)).toBeVisible();
  await expect(page.getByText(/license validation: valid \/ pro/i)).toBeVisible();
  await page.getByRole("button", { name: /enable debug mode/i }).click();
  await expect(page.getByText(/debug mode: enabled/i)).toBeVisible();
  await page.getByRole("button", { name: /disable debug mode/i }).click();
  await expect(page.getByText(/debug mode: disabled/i)).toBeVisible();
});

test("wizard happy path completes all 5 steps and unlocks advanced workspace", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: true });
  await openStarterWizard(page);

  await expect(page.getByText(/step 1 of 5/i)).toBeVisible();
  await expect(page.getByRole("button", { name: /open advanced workspace/i })).toHaveCount(0);

  await page.getByRole("button", { name: /start setup/i }).click();
  await expect(page.getByText(/step 2 of 5/i)).toBeVisible();
  await page.getByRole("button", { name: /complete identity/i }).click();
  await expect(page.getByText(/auth status: registration_finished/i)).toBeVisible();
  await page.getByRole("button", { name: /next step/i }).click();

  await expect(page.getByText(/step 3 of 5/i)).toBeVisible();
  await page.getByRole("button", { name: /load sample data \(recommended\)/i }).click();
  await expect(page.getByText(/onboarding sample data (loaded|already available)/i)).toBeVisible();
  await page.getByRole("button", { name: /next step/i }).click();

  await expect(page.getByText(/step 4 of 5/i)).toBeVisible();
  await page.getByLabel(/part number/i).fill("PN-900");
  await page.getByLabel(/item title/i).fill("Wizard Item");
  await page.getByLabel(/^brand$/i).fill("AFX");
  await page.getByRole("button", { name: /add first item/i }).click();
  await expect(page.getByText(/step 5 of 5/i)).toBeVisible();

  await page.getByLabel(/onboarding theme/i).selectOption("dark");
  await page.getByLabel(/onboarding backup frequency/i).selectOption("weekly");
  await page.getByLabel(/onboarding scanner schedule/i).selectOption("weekly");
  await page.getByRole("button", { name: /finish onboarding/i }).click();

  await expect(page.getByRole("heading", { name: /discovery scanner/i })).toBeVisible();
  const completion = await page.evaluate(() => ({
    completed: localStorage.getItem("cabinet.onboarding.completed.p1"),
    workspace: localStorage.getItem("cabinet.workspace.p1"),
  }));
  expect(completion.completed).toBe("1");
  expect(completion.workspace).toBe("1");
});

test("wizard supports skip path for optional starter-data choices", async ({ page }) => {
  const state = await installWizardMocks(page, { requiresRegistration: true });
  await openStarterWizard(page);

  await page.getByRole("button", { name: /start setup/i }).click();
  await page.getByRole("button", { name: /complete identity/i }).click();
  await page.getByRole("button", { name: /next step/i }).click();

  await expect(page.getByText(/step 3 of 5/i)).toBeVisible();
  await page.getByRole("button", { name: /start empty/i }).click();
  await expect(page.getByText(/starting with an empty collection/i)).toBeVisible();
  await page.getByRole("button", { name: /next step/i }).click();

  await page.getByLabel(/part number/i).fill("PN-901");
  await page.getByLabel(/item title/i).fill("Minimal Item");
  await page.getByRole("button", { name: /add first item/i }).click();
  await page.getByRole("button", { name: /finish onboarding/i }).click();

  await expect(page.getByRole("heading", { name: /discovery scanner/i })).toBeVisible();
  expect(state.sampleCalls).toBeGreaterThanOrEqual(1);
});

test("wizard resume path restores in-progress step after app reload", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: true });
  await openStarterWizard(page);

  await page.getByRole("button", { name: /start setup/i }).click();
  await page.getByRole("button", { name: /complete identity/i }).click();
  await page.getByRole("button", { name: /next step/i }).click();
  await expect(page.getByText(/step 3 of 5/i)).toBeVisible();

  await page.reload();
  await page.getByRole("button", { name: /use default/i }).click();
  await expect(page.getByText(/step 3 of 5/i)).toBeVisible();
  await expect(page.getByText(/setup path: quick/i)).toBeVisible();
});

test("wizard failure path blocks progress then allows retry after identity error", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: true, registerBeginFailsOnce: true });
  await openStarterWizard(page);

  await page.getByRole("button", { name: /start setup/i }).click();
  await expect(page.getByText(/step 2 of 5/i)).toBeVisible();

  const nextStep = page.getByRole("button", { name: /next step/i });
  await expect(nextStep).toBeDisabled();
  await page.getByRole("button", { name: /complete identity/i }).click();
  await expect(page.getByText(/request_failed:\/api\/auth\/webauthn\/register\/begin/i)).toBeVisible();
  await expect(nextStep).toBeDisabled();

  await page.getByRole("button", { name: /complete identity/i }).click();
  await expect(page.getByText(/auth status: registration_finished/i)).toBeVisible();
  await expect(nextStep).toBeEnabled();
});
