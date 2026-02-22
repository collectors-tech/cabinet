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
});
