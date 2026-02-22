import { expect, test } from "@playwright/test";

type MockState = {
  items: Array<{ id: string; part_number: string; title: string; brand?: string; category?: string; series?: string }>;
  savedFilters: Array<{ id: string; name: string; query: Record<string, unknown> }>;
  photos: Array<{ id: string; item_id: string; filename: string; is_primary?: boolean }>;
  querySets: Array<{ id: string; name: string }>;
  candidates: Array<{ id: string; title?: string; listing_id?: string; status?: string }>;
  wishlist: Array<{ id: string; item_id: string; target_price?: number; below_target_now?: boolean; priority?: string }>;
};

async function installUIMocks(page: Parameters<typeof test>[0]["page"]) {
  const state: MockState = {
    items: [{ id: "i1", part_number: "PN-001", title: "Existing Item", brand: "AFX", category: "Cars", series: "Main" }],
    savedFilters: [],
    photos: [{ id: "ph1", item_id: "i1", filename: "one.jpg", is_primary: true }],
    querySets: [{ id: "q1", name: "AFX Query" }],
    candidates: [{ id: "c1", title: "AFX P-1", listing_id: "L1", status: "new" }],
    wishlist: [{ id: "w1", item_id: "i1", target_price: 20, below_target_now: true, priority: "high" }],
  };

  await page.route("**/api/**", async (route) => {
    const req = route.request();
    const method = req.method();
    const url = new URL(req.url());
    const path = url.pathname;
    const qp = url.searchParams;
    const bodyText = req.postData() ?? "";
    const json = bodyText ? safeJSON(bodyText) : {};

    const respondJSON = (payload: unknown, status = 200) =>
      route.fulfill({ status, contentType: "application/json", body: JSON.stringify(payload) });

    if (path === "/api/profiles" && method === "GET") {
      return respondJSON({ profiles: [{ id: "p1", name: "Alpha" }] });
    }
    if (path === "/api/profiles/active" && method === "PUT") {
      return respondJSON({ id: "p1", name: "Alpha" });
    }
    if (path === "/api/profiles/p1/storage" && method === "GET") {
      return respondJSON({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" });
    }
    if (path === "/api/auth/requirements" && method === "GET") {
      return respondJSON({ requires_registration: false });
    }

    if (path === "/api/items" && method === "GET") {
      return respondJSON({ items: state.items });
    }
    if (path === "/api/items" && method === "POST") {
      const next = {
        id: `i${state.items.length + 1}`,
        part_number: String((json as Record<string, unknown>).part_number ?? ""),
        title: String((json as Record<string, unknown>).title ?? ""),
        brand: String((json as Record<string, unknown>).brand ?? ""),
        category: String((json as Record<string, unknown>).category ?? ""),
      };
      state.items.push(next);
      return respondJSON(next, 201);
    }
    if (path === "/api/search/items" && method === "GET") {
      const q = (qp.get("q") || "").toLowerCase();
      const brand = qp.get("brand") || "";
      const filtered = state.items.filter((i) => {
        const qMatch = !q || i.title.toLowerCase().includes(q) || i.part_number.toLowerCase().includes(q);
        const bMatch = !brand || i.brand === brand;
        return qMatch && bMatch;
      });
      return respondJSON({ items: filtered });
    }
    if (path === "/api/profiles/p1/saved-filters" && method === "GET") {
      return respondJSON({ saved_filters: state.savedFilters });
    }
    if (path === "/api/profiles/p1/saved-filters" && method === "POST") {
      const next = {
        id: `f${state.savedFilters.length + 1}`,
        name: String((json as Record<string, unknown>).name ?? "Filter"),
        query: ((json as Record<string, unknown>).query as Record<string, unknown>) ?? {},
      };
      state.savedFilters.push(next);
      return respondJSON(next, 201);
    }

    if (path === "/api/auth/webauthn/register/begin" && method === "POST") {
      return respondJSON({ session_id: "sess-register", options: {} });
    }
    if (path === "/api/auth/webauthn/login/begin" && method === "POST") {
      return respondJSON({ session_id: "sess-login", options: {} });
    }
    if (path === "/api/auth/webauthn/register/finish" && method === "POST") {
      return respondJSON({ credential_id: "cred-1" });
    }
    if (path === "/api/auth/webauthn/login/finish" && method === "POST") {
      return respondJSON({ session_token: "token-1" });
    }
    if (path === "/api/auth/session/validate" && method === "POST") {
      return respondJSON({ valid: true });
    }
    if (path === "/api/auth/session/lock" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/auth/recovery/passphrase" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/auth/recovery/reset/begin" && method === "POST") {
      return respondJSON({ session_id: "sess-recovery" });
    }

    if (path === "/api/items/i1/photos" && method === "GET") {
      return respondJSON({ photos: state.photos.filter((p) => p.item_id === "i1") });
    }
    if (path === "/api/items/i1/photos" && method === "POST") {
      state.photos.push({ id: `ph${state.photos.length + 1}`, item_id: "i1", filename: "upload.jpg", is_primary: false });
      return respondJSON({ ok: true }, 201);
    }
    if (path.startsWith("/api/items/i1/photos/") && path.endsWith("/primary") && method === "PUT") {
      const photoID = path.split("/")[5];
      state.photos = state.photos.map((p) => ({ ...p, is_primary: p.id === photoID }));
      return route.fulfill({ status: 204 });
    }
    if (path.startsWith("/api/items/i1/photos/") && method === "DELETE") {
      const photoID = path.split("/")[5];
      state.photos = state.photos.filter((p) => p.id !== photoID);
      return route.fulfill({ status: 204 });
    }
    if (path === "/api/items/i1/photos-rebuild" && method === "POST") {
      return respondJSON({ ok: true });
    }

    if (path === "/api/scanner/query-sets" && method === "GET") {
      return respondJSON({ query_sets: state.querySets });
    }
    if (path === "/api/scanner/query-sets" && method === "POST") {
      const next = { id: `q${state.querySets.length + 1}`, name: String((json as Record<string, unknown>).name ?? "Query Set") };
      state.querySets.push(next);
      return respondJSON(next, 201);
    }
    if (path === "/api/scanner/run" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/scanner/candidates" && method === "GET") {
      return respondJSON({ candidates: state.candidates });
    }
    if (path === "/api/matching/results" && method === "GET") {
      return respondJSON({
        results: [
          { candidate_id: "c1", state: "matched", part_number: "PN-001" },
          { candidate_id: "c2", state: "suggested", part_number: "PN-009" },
          { candidate_id: "c3", state: "not_in_collection", part_number: "PN-777" },
        ],
      });
    }
    if (path === "/api/discovery/not-in-collection" && method === "GET") {
      return respondJSON({ items: [{ candidate_id: "c3", title: "Unowned Item", price: 12, url: "http://example/listing/c3", last_seen: "2026-02-22" }] });
    }
    if (path === "/api/discovery/action" && method === "POST") {
      return respondJSON({ ok: true });
    }

    if (path === "/api/dashboard" && method === "GET") {
      return respondJSON({ new_discoveries: 1, wishlist_hits: 1, price_drops: 1, total_items: 10, total_instances: 10 });
    }
    if (path === "/api/wishlist" && method === "GET") {
      return respondJSON({ wishlist: state.wishlist });
    }
    if (path === "/api/wishlist" && method === "POST") {
      state.wishlist.push({ id: `w${state.wishlist.length + 1}`, item_id: "i1", target_price: 15, below_target_now: false, priority: "normal" });
      return respondJSON({ ok: true }, 201);
    }
    if (path === "/api/wishlist" && method === "DELETE") {
      const id = qp.get("id") || "";
      state.wishlist = state.wishlist.filter((w) => w.id !== id);
      return route.fulfill({ status: 204 });
    }
    if (path === "/api/pricing/graph" && method === "GET") {
      return respondJSON({ points: [{ day: "2026-02-20", price: 20 }, { day: "2026-02-21", price: 18 }] });
    }
    if (path === "/api/pricing/by-source" && method === "GET") {
      return respondJSON({ by_source: { ebay: [{ snapshot_date: "2026-02-21", min_price: 10, median_price: 11, latest_price: 12 }] } });
    }
    if (path === "/api/pricing/history/export" && method === "GET") {
      return route.fulfill({ status: 200, contentType: "text/csv", body: "date,price\n2026-02-21,18" });
    }

    if (path === "/api/profiles/p1/settings" && method === "GET") {
      return respondJSON({ settings: { scanner_schedule: "0 6 * * *", backup_frequency: "daily", "storage.db_path": "/tmp/p1.db" } });
    }
    if (path === "/api/profiles/p1/settings" && method === "PUT") {
      return respondJSON({ settings: (json as Record<string, unknown>).settings ?? {} });
    }
    if (path === "/api/profiles/p1/secrets" && method === "PUT") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/settings/reset-ignore-rules" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/license/status" && method === "GET") {
      return respondJSON({ state: "valid", tier: "pro" });
    }
    if (path === "/api/logs/activity" && method === "GET") {
      return respondJSON({ activity: [{ event: "scanner_run_completed" }] });
    }
    if (path === "/api/logs/export" && method === "GET") {
      return route.fulfill({ status: 200, contentType: "text/plain", body: "log data" });
    }
    if (path === "/api/data/export/json" && method === "GET") {
      return route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
    }
    if (path === "/api/data/reindex" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/data/repair" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/backup/run" && method === "POST") {
      return respondJSON({ ok: true });
    }

    if (path === "/api/items/i1/barcodes" && method === "GET") {
      return respondJSON({ barcodes: [{ id: "b1", barcode: "12345" }] });
    }
    if (path === "/api/items/i1/barcodes" && method === "POST") {
      return respondJSON({ id: "b2", barcode: "67890" }, 201);
    }
    if (path === "/api/barcodes/12345" && method === "GET") {
      return respondJSON({ matches: [{ item_id: "i1", part_number: "PN-001" }] });
    }
    if (path === "/api/barcodes/12345/external-search" && method === "GET") {
      return respondJSON({ source: "ebay", url: "https://www.ebay.com/sch/i.html?_nkw=12345" });
    }

    if (path === "/api/ai/toggle" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/ai/test" && method === "POST") {
      return respondJSON({ ok: true });
    }
    if (path === "/api/ai/suggest/title" && method === "POST") {
      return respondJSON({ title: "Suggested Title", confidence: 0.9 });
    }
    if (path === "/api/ai/suggest/photo" && method === "POST") {
      return respondJSON({ title: "Photo Suggested", confidence: 0.8 });
    }

    return respondJSON({});
  });
}

function safeJSON(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return {};
  }
}

async function bootApp(page: Parameters<typeof test>[0]["page"]) {
  await installUIMocks(page);
  await page.goto("/");
  await page.getByRole("button", { name: /use alpha/i }).click();
  await expect(page.getByText(/active profile: alpha/i)).toBeVisible();
}

test("api health smoke (real backend)", async ({ request }) => {
  const health = await request.get("/healthz");
  expect(health.ok()).toBeTruthy();
});

test("auth webauthn and session lock flow", async ({ page }) => {
  await bootApp(page);
  await page.getByRole("button", { name: /begin webauthn registration/i }).click();
  await expect(page.getByText(/auth session: sess-register/i)).toBeVisible();
  await page.getByRole("button", { name: /begin webauthn login/i }).click();
  await expect(page.getByText(/auth session: sess-login/i)).toBeVisible();
  await page.getByRole("button", { name: /finish login/i }).click();
  await page.getByRole("button", { name: /validate session/i }).click();
  await page.getByRole("button", { name: /lock session/i }).click();
  await expect(page.getByText(/auth status: session_locked/i)).toBeVisible();
});

test("collection add/filter/saved view and column browsing", async ({ page }) => {
  await bootApp(page);
  await page.getByLabel(/part number/i).fill("PN-002");
  await page.getByLabel(/item title/i).fill("New Item");
  await page.getByLabel(/^brand$/i).fill("Hot Wheels");
  await page.getByRole("button", { name: /add item/i }).click();
  await expect(page.getByRole("cell", { name: /pn-002/i })).toBeVisible();

  await page.getByLabel(/collection search/i).fill("New");
  await page.getByLabel(/saved filter name/i).fill("My Filter");
  await page.getByRole("button", { name: /save current filter/i }).click();
  await page.getByRole("button", { name: /load saved filters/i }).click();
  await expect(page.getByRole("button", { name: /my filter/i })).toBeVisible();

  await page.getByRole("button", { name: /hot wheels/i }).click();
  await expect(page.getByText(/items/i)).toBeVisible();
});

test("photos fullscreen, discovery actions, pricing/wishlist/settings, barcode, and ai", async ({ page }) => {
  await bootApp(page);

  await page.getByRole("button", { name: /load photos/i }).click();
  await page.getByRole("button", { name: /open fullscreen preview/i }).first().click();
  await expect(page.getByText(/fullscreen: one.jpg/i)).toBeVisible();
  await page.getByRole("button", { name: /close fullscreen/i }).click();

  await page.getByRole("button", { name: /load matching results/i }).click();
  await expect(page.getByText(/matched: 1/i)).toBeVisible();
  await page.getByRole("button", { name: /load not in collection/i }).click();
  await page.getByRole("button", { name: /create item/i }).first().click();

  await page.getByRole("button", { name: /load wishlist/i }).click();
  await expect(page.getByText(/below target/i)).toBeVisible();
  await page.getByRole("button", { name: /load pricing sources/i }).click();
  await expect(page.getByText(/source groups: 1/i)).toBeVisible();
  await page.getByRole("button", { name: /export pricing history/i }).click();
  await expect(page.getByText(/export bytes:/i)).toBeVisible();

  await page.getByRole("button", { name: /load profile settings/i }).click();
  await page.getByRole("button", { name: /save profile settings/i }).click();
  await page.getByLabel(/openai api key/i).fill("sk-test");
  await page.getByRole("button", { name: /save openai key/i }).click();
  await expect(page.getByText(/openai_key_saved/i)).toBeVisible();

  await page.getByRole("button", { name: /load barcodes/i }).click();
  await page.getByRole("button", { name: /lookup barcode/i }).click();
  await expect(page.getByText(/local matches: 1/i)).toBeVisible();

  page.on("dialog", (dialog) => dialog.accept());
  await page.getByRole("button", { name: /enable ai/i }).click();
  await page.getByLabel(/ai title input/i).fill("AFX listing");
  await page.getByRole("button", { name: /suggest from title/i }).click();
  await page.getByRole("button", { name: /apply suggestion/i }).click();
  await expect(page.getByText(/ai confidence:/i)).toBeVisible();
});
