import { expect, test } from "@playwright/test";

type WizardMockOptions = {
  registerBeginFailsOnce?: boolean;
  sampleFailsOnce?: boolean;
  requiresRegistration?: boolean;
};

async function installWizardMocks(page: Parameters<typeof test>[0]["page"], options: WizardMockOptions = {}) {
  const state = {
    items: [
      { id: "i1", part_number: "PN-001", title: "Existing Item", brand: "AFX", category: "Cars" },
      { id: "i2", part_number: "PN-002", title: "Second Item", brand: "Tyco", category: "Cars" },
    ] as Array<{ id: string; part_number: string; title: string; brand?: string; category?: string }>,
    photosByItem: {
      i1: [
        { id: "ph1", item_id: "i1", filename: "a.jpg", is_primary: true },
        { id: "ph2", item_id: "i1", filename: "b.jpg", is_primary: false },
      ],
    } as Record<string, Array<{ id: string; item_id: string; filename: string; is_primary?: boolean }>>,
    nextPhotoNumber: 3,
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
    let body: Record<string, unknown> | undefined;
    try {
      body = req.postDataJSON?.() as Record<string, unknown> | undefined;
    } catch {
      body = undefined;
    }
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
    if (path === "/api/search/items" && method === "GET") {
      const q = String(url.searchParams.get("q") || "").toLowerCase();
      const brand = String(url.searchParams.get("brand") || "").toLowerCase();
      const category = String(url.searchParams.get("category") || "").toLowerCase();
      const sort = String(url.searchParams.get("sort") || "date_added").toLowerCase();
      let filtered = state.items.filter((item) => {
        const matchesQ =
          !q ||
          item.part_number.toLowerCase().includes(q) ||
          item.title.toLowerCase().includes(q) ||
          String(item.brand || "").toLowerCase().includes(q);
        const matchesBrand = !brand || String(item.brand || "").toLowerCase().includes(brand);
        const matchesCategory = !category || String(item.category || "").toLowerCase().includes(category);
        return matchesQ && matchesBrand && matchesCategory;
      });
      if (sort === "part_number") {
        filtered = [...filtered].sort((a, b) => a.part_number.localeCompare(b.part_number));
      }
      return respondJSON({ items: filtered });
    }
    if (path === "/api/items/bulk-edit" && method === "POST") {
      const itemIDs = Array.isArray(body?.item_ids) ? (body?.item_ids as string[]) : [];
      const changes = (body?.changes || {}) as Record<string, unknown>;
      state.items = state.items.map((item) =>
        itemIDs.includes(item.id)
          ? {
              ...item,
              brand: typeof changes.brand === "string" ? changes.brand : item.brand,
              category: typeof changes.category === "string" ? changes.category : item.category,
            }
          : item,
      );
      return respondJSON({ updated_count: itemIDs.length });
    }
    if (path.startsWith("/api/items/") && method === "PUT" && !path.includes("/photos/")) {
      const itemID = path.split("/")[3] || "";
      state.items = state.items.map((item) =>
        item.id === itemID
          ? {
              ...item,
              title: typeof body?.title === "string" ? body.title : item.title,
              brand: typeof body?.brand === "string" ? body.brand : item.brand,
            }
          : item,
      );
      const updated = state.items.find((item) => item.id === itemID);
      return respondJSON(updated || { id: itemID, title: String(body?.title || ""), brand: String(body?.brand || "") });
    }
    if (path.startsWith("/api/items/") && path.endsWith("/photos") && method === "GET") {
      const itemID = path.split("/")[3] || "";
      return respondJSON({ photos: state.photosByItem[itemID] || [] });
    }
    if (path.startsWith("/api/items/") && path.endsWith("/photos") && method === "POST") {
      const itemID = path.split("/")[3] || "";
      const next = {
        id: `ph${state.nextPhotoNumber}`,
        item_id: itemID,
        filename: `uploaded-${state.nextPhotoNumber}.jpg`,
      };
      state.nextPhotoNumber += 1;
      state.photosByItem[itemID] = [...(state.photosByItem[itemID] || []), next];
      return respondJSON(next, 201);
    }
    if (path.startsWith("/api/items/") && path.endsWith("/photos/reorder") && method === "POST") {
      const itemID = path.split("/")[3] || "";
      const nextOrder = Array.isArray(body?.photo_ids) ? (body.photo_ids as string[]) : [];
      const existing = state.photosByItem[itemID] || [];
      const byID = new Map(existing.map((photo) => [photo.id, photo]));
      state.photosByItem[itemID] = nextOrder.map((photoID) => byID.get(photoID)).filter(Boolean) as Array<{
        id: string;
        item_id: string;
        filename: string;
        is_primary?: boolean;
      }>;
      return respondJSON({ ok: true });
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

test("collection bulk edit, inline edit, photo staging, and shell scroll ownership", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: false });
  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "1");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();

  const shellStyles = await page.evaluate(() => {
    const nav = document.querySelector("aside.primary-nav");
    const header = document.querySelector("header.page-header");
    const content = document.querySelector(".cabinet-content");
    const contentScroll = document.querySelector(".cabinet-content-scroll");
    if (!nav || !header || !content || !contentScroll) {
      return { ready: false };
    }
    return {
      ready: true,
      navOverflow: getComputedStyle(nav).overflowY,
      headerPosition: getComputedStyle(header).position,
      contentOverflow: getComputedStyle(content).overflowY,
      scrollOverflow: getComputedStyle(contentScroll).overflowY,
    };
  });
  expect(shellStyles.ready).toBe(true);
  expect(shellStyles.headerPosition).toBe("sticky");
  expect(shellStyles.contentOverflow).toBe("hidden");
  expect(shellStyles.scrollOverflow).toBe("auto");

  await page.getByRole("button", { name: /^collection$/i }).first().click();
  await page.getByRole("checkbox", { name: /select item pn-001/i }).check();
  await page.getByRole("checkbox", { name: /select item pn-002/i }).check();
  await page.getByLabel(/bulk edit brand/i).fill("Auto World");
  await page.getByRole("button", { name: /preview bulk edit/i }).click();
  await expect(page.getByRole("button", { name: /apply bulk edit/i })).toBeDisabled();
  await page.getByRole("checkbox", { name: /confirm bulk edit changes/i }).check();
  await expect(page.getByRole("button", { name: /apply bulk edit/i })).toBeEnabled();
  await page.getByRole("button", { name: /apply bulk edit/i }).click();
  await expect(page.getByText(/bulk_edit_applied:2/i)).toBeVisible();

  await page.getByLabel(/inline title pn-001/i).fill("Updated Item One");
  await page.getByRole("button", { name: /save inline/i }).first().click();
  await expect(page.getByText(/inline_updated:i1/i)).toBeVisible();
  await expect(page.getByRole("cell", { name: "Updated Item One", exact: true })).toBeVisible();

  await page.getByRole("button", { name: /^photos$/i }).first().click();
  await page.getByRole("button", { name: /load photos/i }).click();
  const files = [
    { name: "upload-a.jpg", mimeType: "image/jpeg", buffer: Buffer.from("a") },
    { name: "upload-b.jpg", mimeType: "image/jpeg", buffer: Buffer.from("b") },
  ];
  await page.getByLabel(/photo files/i).setInputFiles(files);
  await expect(page.getByText(/staged files: 2/i)).toBeVisible();
  await page.getByRole("button", { name: /upload staged photos/i }).click();
  await expect(page.getByText(/staged files: 0/i)).toBeVisible();

  const firstBefore = (await page.getByRole("button", { name: /move down/i }).first().locator("xpath=ancestor::li[1]").innerText()).toLowerCase();
  await page.getByRole("button", { name: /move down/i }).first().click();
  const firstAfter = (await page.getByRole("button", { name: /move down/i }).first().locator("xpath=ancestor::li[1]").innerText()).toLowerCase();
  expect(firstAfter).not.toBe(firstBefore);
});

test("collection command bar supports summarize toggle and view controls", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: false });
  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "1");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();

  await page.getByRole("button", { name: /^collection$/i }).first().click();
  await expect(page.getByRole("region", { name: /collection command bar/i })).toBeVisible();
  await expect(page.getByText(/summary mode: detailed/i)).toBeVisible();

  await page.getByLabel(/summarize items/i).check();
  await expect(page.getByText(/summary mode: summarized/i)).toBeVisible();

  await page.getByRole("button", { name: /table view/i }).click();
  await expect(page.getByText(/view mode: table/i)).toBeVisible();
  await page.getByRole("button", { name: /card view/i }).click();
  await expect(page.getByText(/view mode: cards/i)).toBeVisible();
});

test("three-pane shell renders context pane and keeps context in mobile drawer", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: false });
  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "1");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();

  await expect(page.locator("aside.primary-nav")).toBeVisible();
  await expect(page.getByLabel("Collection context pane").first()).toBeVisible();
  await expect(page.getByLabel("Primary content")).toBeVisible();
  await page.getByRole("button", { name: /wishlist focus/i }).click();
  await expect(page.getByText(/context: wishlist focus/i)).toBeVisible();

  await page.getByRole("button", { name: /collapse collection pane/i }).click();
  await expect(page.getByRole("button", { name: /expand collection pane/i })).toBeVisible();

  await page.setViewportSize({ width: 390, height: 844 });
  await page.getByRole("button", { name: /open navigation menu/i }).click();
  const drawer = page.getByRole("dialog", { name: /navigation menu/i });
  await expect(drawer).toBeVisible();
  await expect(drawer.getByLabel("Collection context pane")).toBeVisible();
});

test("chat rail supports context chips and preview-confirm workspace actions", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: false });
  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "1");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();

  await page.getByRole("button", { name: /toggle chat copilot/i }).click();
  const chatRail = page.getByRole("complementary", { name: /chat copilot/i });
  await expect(chatRail).toBeVisible();
  const collectionNav = page.getByRole("button", { name: /^collection$/i }).first();
  await expect(collectionNav).not.toHaveAttribute("aria-current", "page");

  await chatRail.getByRole("button", { name: /wishlist hits context chip/i }).click();
  await expect(chatRail.getByLabel(/chat message/i)).toHaveValue(/wishlist hits/i);

  await chatRail.getByRole("button", { name: /preview open collection workspace action/i }).click();
  await expect(chatRail.getByText(/ready to apply: open collection workspace/i)).toBeVisible();
  await expect(collectionNav).not.toHaveAttribute("aria-current", "page");
  await chatRail.getByRole("button", { name: /confirm apply action/i }).click();
  await expect(collectionNav).toHaveAttribute("aria-current", "page");
  await expect(page.getByRole("heading", { name: /^collection$/i })).toBeVisible();
});

test("chat open-close preserves active workspace and supports local attachment staging", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: false });
  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "1");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();
  await page.getByRole("button", { name: /^scanner$/i }).first().click();
  await expect(page.getByRole("heading", { name: /discovery scanner/i })).toBeVisible();

  await page.getByRole("button", { name: /toggle chat copilot/i }).click();
  const chatRail = page.getByRole("complementary", { name: /chat copilot/i });
  await expect(chatRail).toBeVisible();

  await chatRail.getByLabel(/chat attachment/i).setInputFiles({
    name: "notes.txt",
    mimeType: "text/plain",
    buffer: Buffer.from("collector notes"),
  });
  await expect(chatRail.getByText(/notes\.txt/i)).toBeVisible();
  await chatRail.getByRole("button", { name: /remove attachment notes\.txt/i }).click();
  await expect(chatRail.getByText(/notes\.txt/i)).toHaveCount(0);

  await chatRail.getByLabel(/chat message/i).fill("show collection gaps");
  await chatRail.getByRole("button", { name: /^send$/i }).click();
  await expect(chatRail.getByText(/you:\s*show collection gaps/i)).toBeVisible();

  await chatRail.getByRole("button", { name: /close chat copilot/i }).click();
  await expect(page.getByRole("complementary", { name: /chat copilot/i })).toHaveCount(0);
  await expect(page.getByRole("heading", { name: /discovery scanner/i })).toBeVisible();
});

test("wizard happy path completes all 5 steps and unlocks advanced workspace", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: true });
  await openStarterWizard(page);

  await expect(page.getByText(/step 1 of 5/i)).toBeVisible();
  await expect(page.getByRole("button", { name: /open advanced workspace/i })).toHaveCount(0);

  await page.getByRole("button", { name: /start setup/i }).click();
  await expect(page.getByText(/step 2 of 5/i)).toBeVisible();
  await page.getByRole("button", { name: /complete identity/i }).click();
  await expect(page.locator(".cabinet-onboarding").getByText(/identity complete: yes/i)).toBeVisible();
  await page.getByRole("button", { name: /next step/i }).click();

  await expect(page.getByText(/step 3 of 5/i)).toBeVisible();
  await page.getByRole("button", { name: /load sample data \(recommended\)/i }).click();
  await expect(page.locator(".cabinet-onboarding").getByText(/starter data: sample/i)).toBeVisible();
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

test("home screen adapts for first-run and returning users", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: true });

  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "0");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();
  await expect(page.getByRole("heading", { name: /starter onboarding wizard/i })).toBeVisible();
  await expect(page.getByText(/onboarding complete\./i)).toHaveCount(0);

  await page.addInitScript(() => {
    localStorage.setItem("cabinet.workspace.p1", "1");
    localStorage.setItem("cabinet.onboarding.completed.p1", "1");
  });
  await page.reload();
  await page.getByRole("button", { name: /use default/i }).click();
  await expect(page.getByText(/onboarding complete\./i)).toBeVisible();
  await expect(page.getByRole("heading", { name: /starter onboarding wizard/i })).toHaveCount(0);
  await expect(page.locator(".cabinet-home-diagnostics")).not.toHaveAttribute("open", "");
});

test("home attention cards support dismiss and snooze persistence", async ({ page }) => {
  await installWizardMocks(page, { requiresRegistration: false });
  await page.addInitScript(() => {
    localStorage.clear();
    localStorage.setItem("cabinet.workspace.p1", "1");
    localStorage.removeItem("cabinet.attention.dismissed.p1");
    localStorage.removeItem("cabinet.attention.snoozed.p1");
  });
  await page.goto("/");
  await page.getByRole("button", { name: /use default/i }).click();
  await page.getByRole("button", { name: /^dashboard$/i }).first().click();

  await expect(page.getByRole("button", { name: /review discoveries/i })).toBeVisible();
  await page.getByRole("button", { name: /dismiss review new discoveries/i }).click();
  await expect(page.getByRole("button", { name: /review discoveries/i })).toHaveCount(0);

  const dismissedStorage = await page.evaluate(() => localStorage.getItem("cabinet.attention.dismissed.p1") || "");
  expect(dismissedStorage).toContain("discoveries");

  await page.getByRole("button", { name: /snooze review wishlist hits/i }).click();
  const snoozedStorage = await page.evaluate(() => localStorage.getItem("cabinet.attention.snoozed.p1") || "");
  expect(snoozedStorage).toContain("wishlist");
});

test("wizard supports skip path for optional starter-data choices", async ({ page }) => {
  const state = await installWizardMocks(page, { requiresRegistration: true });
  await openStarterWizard(page);

  await page.getByRole("button", { name: /start setup/i }).click();
  await page.getByRole("button", { name: /complete identity/i }).click();
  await page.getByRole("button", { name: /next step/i }).click();

  await expect(page.getByText(/step 3 of 5/i)).toBeVisible();
  await page.getByRole("button", { name: /start empty/i }).click();
  await expect(page.locator(".cabinet-onboarding").getByText(/starter data: empty/i)).toBeVisible();
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
  await expect(page.locator(".cabinet-onboarding").getByText(/identity complete: yes/i)).toBeVisible();
  await expect(nextStep).toBeEnabled();
});
