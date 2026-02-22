import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

describe("App shell", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders shell and theme control", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);
    expect(screen.getByTestId("app-shell")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /toggle theme/i })).toBeInTheDocument();
  });

  it("shows onboarding create flow when no profiles exist", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ profiles: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 201 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ db_path: "C:/Cabinet/profiles/p1/cabinet.db", media_dir: "C:/Cabinet/profiles/p1/media" }), {
          status: 200,
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const create = await screen.findByRole("button", { name: /create first profile/i });
    create.click();
    expect(await screen.findByText(/active profile: default/i)).toBeInTheDocument();
  });

  it("allows activating an existing profile", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }, { id: "p2", name: "Beta" }] }), {
          status: 200,
        });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p2", name: "Beta" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p2/storage")) {
        return new Response(JSON.stringify({ db_path: "/tmp/p2.db", media_dir: "/tmp/p2/media" }), { status: 200 });
      }
      if (url.includes("/api/auth/requirements?profile_id=p2")) {
        return new Response(JSON.stringify({ requires_registration: false }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use beta/i });
    activate.click();
    expect(await screen.findByText(/active profile: beta/i)).toBeInTheDocument();
    expect(await screen.findByText(/database: \/tmp\/p2.db/i)).toBeInTheDocument();
    expect(await screen.findByText(/media: \/tmp\/p2\/media/i)).toBeInTheDocument();
  });

  it("starts WebAuthn registration for active profile", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ profiles: [{ id: "p2", name: "Beta" }] }), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p2", name: "Beta" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ db_path: "/tmp/p2.db", media_dir: "/tmp/p2/media" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requires_registration: true }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ session_id: "sess-reg-1", options: {} }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use beta/i });
    activate.click();

    const begin = await screen.findByRole("button", { name: /begin webauthn registration/i });
    begin.click();
    expect(await screen.findByText(/auth session: sess-reg-1/i)).toBeInTheDocument();
  });

  it("lists and creates collection items", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p1", name: "Alpha" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requires_registration: false }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-001", title: "Existing", brand: "AFX" }] }), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: "i2", part_number: "PN-002", title: "New Item", brand: "Hot Wheels" }), {
          status: 201,
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    expect(await screen.findByText(/pn-001/i)).toBeInTheDocument();

    const partInput = await screen.findByLabelText(/part number/i);
    const titleInput = await screen.findByLabelText(/item title/i);
    const brandInput = await screen.findByLabelText(/brand/i);
    const addButton = await screen.findByRole("button", { name: /add item/i });

    fireEvent.change(partInput, { target: { value: "PN-002" } });
    fireEvent.change(titleInput, { target: { value: "New Item" } });
    fireEvent.change(brandInput, { target: { value: "Hot Wheels" } });
    addButton.click();

    expect(await screen.findByText(/pn-002/i)).toBeInTheDocument();
  });

  it("loads photos for selected item", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p1", name: "Alpha" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requires_registration: false }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-001", title: "Existing", brand: "AFX" }] }), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ photos: [{ id: "ph1", item_id: "i1", filename: "a.jpg", is_primary: true }] }), {
          status: 200,
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const loadPhotos = await screen.findByRole("button", { name: /load photos/i });
    loadPhotos.click();
    expect(await screen.findByText(/a.jpg/i)).toBeInTheDocument();
  });

  it("loads scanner query sets", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p1", name: "Alpha" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requires_registration: false }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ query_sets: [{ id: "q1", name: "AFX Search" }] }), { status: 200 }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const loadQuerySets = await screen.findByRole("button", { name: /load query sets/i });
    loadQuerySets.click();
    expect(await screen.findByText(/afx search/i)).toBeInTheDocument();
  });

  it("loads dashboard, wishlist, and pricing graph", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }] }), { status: 200 });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p1", name: "Alpha" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p1/storage")) {
        return new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 });
      }
      if (url.includes("/api/auth/requirements?profile_id=p1")) {
        return new Response(JSON.stringify({ requires_registration: false }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-1", title: "T1" }] }), { status: 200 });
      }
      if (url === "/api/dashboard") {
        return new Response(
          JSON.stringify({ new_discoveries: 3, wishlist_hits: 1, price_drops: 2, total_items: 10, total_instances: 12 }),
          { status: 200 },
        );
      }
      if (url === "/api/wishlist") {
        return new Response(JSON.stringify({ wishlist: [{ id: "w1", item_id: "i1", target_price: 25 }] }), { status: 200 });
      }
      if (url.includes("/api/pricing/graph?item_id=i1")) {
        return new Response(JSON.stringify({ points: [{ day: "2026-02-20", price: 20 }, { day: "2026-02-21", price: 18 }] }), {
          status: 200,
        });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const loadDashboard = await screen.findByRole("button", { name: /load dashboard/i });
    loadDashboard.click();
    expect(await screen.findByText(/new discoveries: 3/i)).toBeInTheDocument();

    const loadWishlist = await screen.findByRole("button", { name: /load wishlist/i });
    loadWishlist.click();
    expect(await screen.findByText(/wishlist item: i1 target 25/i)).toBeInTheDocument();

    const loadGraph = await screen.findByRole("button", { name: /load pricing graph/i });
    loadGraph.click();
    expect(await screen.findByText(/pricing points: 2/i)).toBeInTheDocument();
  });

  it("loads settings admin status and logs", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }] }), { status: 200 });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p1", name: "Alpha" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p1/storage")) {
        return new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 });
      }
      if (url.includes("/api/auth/requirements?profile_id=p1")) {
        return new Response(JSON.stringify({ requires_registration: false }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      if (url.includes("/api/license/status?profile_id=p1")) {
        return new Response(JSON.stringify({ state: "valid", tier: "pro" }), { status: 200 });
      }
      if (url === "/api/logs/activity?limit=10") {
        return new Response(JSON.stringify({ activity: [{ event: "scanner_run_completed" }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const loadAdmin = await screen.findByRole("button", { name: /load admin status/i });
    loadAdmin.click();
    expect(await screen.findByText(/license: valid \/ pro/i)).toBeInTheDocument();
    expect(await screen.findByText(/log entries: 1/i)).toBeInTheDocument();
  });

  it("supports barcode lookup and external search link", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }] }), { status: 200 });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p1", name: "Alpha" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p1/storage")) {
        return new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 });
      }
      if (url.includes("/api/auth/requirements?profile_id=p1")) {
        return new Response(JSON.stringify({ requires_registration: false }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-1", title: "T1" }] }), { status: 200 });
      }
      if (url.includes("/api/items/i1/barcodes") && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ barcodes: [{ id: "b1", barcode: "12345" }] }), { status: 200 });
      }
      if (url.includes("/api/barcodes/12345/external-search")) {
        return new Response(JSON.stringify({ source: "ebay", url: "https://www.ebay.com/sch/i.html?_nkw=12345" }), { status: 200 });
      }
      if (url === "/api/barcodes/12345") {
        return new Response(JSON.stringify({ matches: [{ item_id: "i1", part_number: "PN-1" }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const loadBarcodes = await screen.findByRole("button", { name: /load barcodes/i });
    loadBarcodes.click();
    expect(await screen.findByText(/12345/i)).toBeInTheDocument();

    const lookupBtn = await screen.findByRole("button", { name: /lookup barcode/i });
    lookupBtn.click();
    expect(await screen.findByText(/local matches: 1/i)).toBeInTheDocument();

    const externalBtn = await screen.findByRole("button", { name: /external search/i });
    externalBtn.click();
    expect(await screen.findByText(/ebay.com\/sch\/i.html/i)).toBeInTheDocument();
  });
});
