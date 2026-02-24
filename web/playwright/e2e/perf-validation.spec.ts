import { expect, test } from "@playwright/test";
import fs from "node:fs";
import path from "node:path";

type PerfResult = {
  dataset: "S2" | "S3";
  initialHomeRenderMs: number;
  navMedianMs: number;
  searchMedianMs: number;
  sortMs: number;
  detailOpenMs: number;
  discoverActionMs: number;
  reportsExportMs: number;
  crashed: boolean;
};

const perfResults: PerfResult[] = [];

function median(values: number[]): number {
  if (values.length === 0) {
    return 0;
  }
  const sorted = [...values].sort((a, b) => a - b);
  const mid = Math.floor(sorted.length / 2);
  if (sorted.length % 2 === 0) {
    return Math.round((sorted[mid-1] + sorted[mid]) / 2);
  }
  return sorted[mid];
}

function buildItems(count: number, prefix: string) {
  return Array.from({ length: count }, (_, idx) => ({
    id: `${prefix}-item-${idx + 1}`,
    part_number: `PN-${prefix}-${String(idx + 1).padStart(6, "0")}`,
    title: `Collector ${idx + 1}`,
    brand: `Brand-${(idx % 60) + 1}`,
    category: `Category-${(idx % 40) + 1}`,
  }));
}

function buildDiscoveries(count: number, prefix: string) {
  return Array.from({ length: count }, (_, idx) => ({
    candidate_id: `${prefix}-cand-${idx + 1}`,
    title: `Discovery ${idx + 1}`,
    price: (idx % 250) + 10,
    stock_state: idx % 7 === 0 ? "low_stock" : idx % 11 === 0 ? "out_of_stock" : "in_stock",
    stock_count: idx % 11 === 0 ? 0 : (idx % 14) + 1,
  }));
}

async function installPerfMocks(
  page: Parameters<typeof test>[0]["page"],
  dataset: "S2" | "S3",
) {
  const itemCount = dataset === "S2" ? 5000 : 25000;
  const discoveryCount = dataset === "S2" ? 2000 : 10000;
  const items = buildItems(itemCount, dataset.toLowerCase());
  const discoveries = buildDiscoveries(discoveryCount, dataset.toLowerCase());
  const profileID = `p-${dataset.toLowerCase()}`;

  await page.route("**/api/**", async (route) => {
    const req = route.request();
    const url = new URL(req.url());
    const pathName = url.pathname;
    const method = req.method();
    const json = (body: unknown, status = 200) =>
      route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });

    if (pathName === "/api/profiles" && method === "GET") {
      return json({ profiles: [{ id: profileID, name: `Perf ${dataset}` }] });
    }
    if (pathName === "/api/profiles/active" && method === "PUT") {
      return json({ id: profileID, name: `Perf ${dataset}` });
    }
    if (pathName === `/api/profiles/${profileID}/storage` && method === "GET") {
      return json({ db_path: "/tmp/perf.db", media_dir: "/tmp/perf-media" });
    }
    if (pathName === "/api/auth/requirements" && method === "GET") {
      return json({ requires_registration: false });
    }
    if (pathName === "/api/dashboard" && method === "GET") {
      return json({
        new_discoveries: discoveryCount,
        wishlist_hits: dataset === "S2" ? 1000 : 5000,
        price_drops: dataset === "S2" ? 600 : 3200,
        recently_added: 120,
        total_items: itemCount,
        total_instances: dataset === "S2" ? 15000 : 80000,
      });
    }
    if (pathName === "/api/items" && method === "GET") {
      const visible = items.slice(0, dataset === "S2" ? 800 : 1200);
      return json({ items: visible });
    }
    if (pathName === "/api/search/items" && method === "GET") {
      const q = (url.searchParams.get("q") || "").toLowerCase();
      const sortBy = (url.searchParams.get("sort") || "date_added").toLowerCase();
      let filtered = items.filter((item) => !q || item.part_number.toLowerCase().includes(q) || item.title.toLowerCase().includes(q));
      if (sortBy === "part_number") {
        filtered = [...filtered].sort((a, b) => a.part_number.localeCompare(b.part_number));
      } else if (sortBy === "price") {
        filtered = [...filtered].sort((a, b) => a.id.localeCompare(b.id));
      }
      return json({ items: filtered.slice(0, dataset === "S2" ? 800 : 1200) });
    }
    if (pathName === "/api/instances" && method === "GET") {
      return json({ instances: [{ id: "inst-1", status: "sealed", condition: "mint", quantity: 1 }] });
    }
    if (pathName === "/api/discovery/not-in-collection" && method === "GET") {
      return json({ items: discoveries.slice(0, dataset === "S2" ? 800 : 1200) });
    }
    if (pathName === "/api/discovery/action" && method === "POST") {
      return json({ ok: true });
    }
    if (pathName === "/api/scanner/run" && method === "POST") {
      return json({ ok: true });
    }
    if (pathName === "/api/wishlist/hits" && method === "GET") {
      return json({ hits: [{ item_id: `${dataset.toLowerCase()}-item-1`, listing_id: "l1", title: "Hit", price: 20 }] });
    }
    if (pathName === "/api/pricing/trend" && method === "GET") {
      return json({ trend: "flat" });
    }
    if (pathName === "/api/pricing/stats" && method === "GET") {
      return json({ min: 10, median: 20, latest: 30 });
    }
    if (pathName === "/api/pricing/by-source" && method === "GET") {
      return json({ by_source: { ebay: [{ snapshot_date: "2026-02-24", min_price: 10, median_price: 20, latest_price: 30, stock_count: 6 }] } });
    }
    if (pathName === "/api/pricing/history/export" && method === "GET") {
      return route.fulfill({ status: 200, contentType: "text/csv", body: "day,min,median,latest\n2026-02-24,10,20,30\n" });
    }
    return json({});
  });

  return { profileID };
}

async function runPerfFlow(page: Parameters<typeof test>[0]["page"], dataset: "S2" | "S3"): Promise<PerfResult> {
  const { profileID } = await installPerfMocks(page, dataset);
  await page.addInitScript((id) => {
    localStorage.clear();
    localStorage.setItem(`cabinet.workspace.${id}`, "1");
  }, profileID);

  let crashed = false;
  page.on("crash", () => {
    crashed = true;
  });

  const initialStart = Date.now();
  await page.goto("/");
  await page.getByRole("button", { name: /use perf/i }).click();
  await page.getByRole("button", { name: /^dashboard$/i }).first().click();
  await expect(page.getByRole("heading", { name: /^dashboard$/i })).toBeVisible();
  const initialHomeRenderMs = Date.now() - initialStart;

  const navTargets: Array<{ name: RegExp; heading: RegExp }> = [
    { name: /^collection$/i, heading: /^collection$/i },
    { name: /^scanner$/i, heading: /discovery scanner/i },
    { name: /^discoveries$/i, heading: /not in my collection/i },
    { name: /^pricing$/i, heading: /^pricing$/i },
    { name: /^reports$/i, heading: /^reports$/i },
    { name: /^settings$/i, heading: /settings and diagnostics/i },
  ];
  const navSamples: number[] = [];
  for (const target of navTargets) {
    const start = Date.now();
    await page.getByRole("button", { name: target.name }).first().click();
    await expect(page.getByRole("heading", { name: target.heading })).toBeVisible();
    navSamples.push(Date.now() - start);
  }

  await page.getByRole("button", { name: /^collection$/i }).first().click();
  await expect(page.getByRole("heading", { name: /^collection$/i })).toBeVisible();

  const searchSamples: number[] = [];
  for (let i = 0; i < 20; i++) {
    const q = `PN-${dataset.toLowerCase()}-${String(i + 1).padStart(6, "0")}`;
    const requestSeen = page.waitForRequest((req) => {
      const url = new URL(req.url());
      return url.pathname === "/api/search/items" && (url.searchParams.get("q") || "").toLowerCase() === q.toLowerCase();
    });
    await page.getByLabel("Collection search").fill(q);
    await requestSeen;
    const start = Date.now();
    await expect(page.locator("table tbody tr").first()).toContainText(q);
    searchSamples.push(Date.now() - start);
  }

  const sortStart = Date.now();
  await page.getByLabel("Collection sort").selectOption("part_number");
  await expect(page.locator("table tbody tr").first()).toBeVisible();
  const sortMs = Date.now() - sortStart;

  const detailStart = Date.now();
  await page.locator("table tbody tr").first().locator("td").nth(1).getByRole("button").click();
  await expect(page.getByLabel("Item ID")).toHaveValue(/-item-/i);
  const detailOpenMs = Date.now() - detailStart;

  await page.getByRole("button", { name: /^discoveries$/i }).first().click();
  const discoverStart = Date.now();
  await page.getByRole("button", { name: /load not in collection/i }).click();
  await expect(page.getByRole("button", { name: /ignore/i }).first()).toBeVisible();
  await page.getByRole("button", { name: /ignore/i }).first().click();
  await expect(page.getByRole("status").first()).toBeVisible();
  const discoverActionMs = Date.now() - discoverStart;

  await page.getByRole("button", { name: /^reports$/i }).first().click();
  const reportStart = Date.now();
  await page.getByLabel("Report item id").fill(`${dataset.toLowerCase()}-item-1`);
  await page.getByRole("button", { name: /load trend summary/i }).click();
  await page.getByRole("button", { name: /load source summary/i }).click();
  await page.getByRole("button", { name: /export report history/i }).click();
  await expect(page.getByText(/export bytes:/i)).toBeVisible();
  const reportsExportMs = Date.now() - reportStart;

  return {
    dataset,
    initialHomeRenderMs,
    navMedianMs: median(navSamples),
    searchMedianMs: median(searchSamples),
    sortMs,
    detailOpenMs,
    discoverActionMs,
    reportsExportMs,
    crashed,
  };
}

test("SCAL-001..005 S2 meets interaction performance targets", async ({ page }) => {
  const result = await runPerfFlow(page, "S2");
  perfResults.push(result);

  expect(result.crashed).toBe(false);
  expect(result.initialHomeRenderMs).toBeLessThanOrEqual(1000);
  expect(result.searchMedianMs).toBeLessThanOrEqual(300);
  expect(result.navMedianMs).toBeLessThanOrEqual(150);
  expect(result.sortMs).toBeLessThanOrEqual(250);
  expect(result.detailOpenMs).toBeLessThanOrEqual(120);
});

test("S3 runs without crash or unrecoverable stall regressions", async ({ page }) => {
  const result = await runPerfFlow(page, "S3");
  perfResults.push(result);
  expect(result.crashed).toBe(false);
  expect(result.discoverActionMs).toBeLessThanOrEqual(2000);
  expect(result.reportsExportMs).toBeLessThanOrEqual(2000);
});

test.afterAll(() => {
  if (perfResults.length === 0) {
    return;
  }
  const byDataset = Object.fromEntries(perfResults.map((result) => [result.dataset, result])) as Record<string, PerfResult>;
  const jsonReportPath = path.resolve(process.cwd(), "..", "docs", "ui-spec", "12-PERF-VALIDATION-S2-S3.json");
  fs.writeFileSync(jsonReportPath, JSON.stringify(byDataset, null, 2));
  const markdownReportPath = path.resolve(process.cwd(), "..", "docs", "ui-spec", "12-PERF-VALIDATION-S2-S3.md");
  const lines: string[] = [
    "# UI Perf Validation (S2/S3)",
    "",
    "| Dataset | Initial Render (ms) | Nav Median (ms) | Search Median (ms) | Sort (ms) | Detail Open (ms) | Discover Action (ms) | Reports Export (ms) | Crashed |",
    "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |",
  ];
  for (const result of perfResults) {
    lines.push(
      `| ${result.dataset} | ${result.initialHomeRenderMs} | ${result.navMedianMs} | ${result.searchMedianMs} | ${result.sortMs} | ${result.detailOpenMs} | ${result.discoverActionMs} | ${result.reportsExportMs} | ${result.crashed ? "yes" : "no"} |`,
    );
  }
  lines.push("");
  lines.push("- Targets:");
  lines.push("  - S2: initial <=1000ms, nav median <=150ms, search median <=300ms, sort <=250ms, detail open <=120ms");
  lines.push("  - S3: no crash; discover action <=2000ms; reports export <=2000ms");
  fs.writeFileSync(markdownReportPath, lines.join("\n"));
});
