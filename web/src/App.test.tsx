import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

describe("App shell", () => {
  afterEach(() => {
    cleanup();
    localStorage.clear();
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

  it("shows API Kitchen Sync quick link in utility links", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);
    const link = await screen.findByRole("link", { name: /api kitchen sync/i });
    expect(link).toHaveAttribute("href", "/apidocs");
  });

  it("opens and closes mobile navigation drawer", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);
    const openButton = await screen.findByRole("button", { name: /open navigation menu/i });
    fireEvent.click(openButton);
    expect(await screen.findByRole("dialog", { name: /navigation menu/i })).toBeInTheDocument();

    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("dialog", { name: /navigation menu/i })).not.toBeInTheDocument();
  });

  it("shows onboarding create flow when no profiles exist", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [] }), { status: 200 });
      }
      if (url === "/api/profiles" && init?.method === "POST") {
        return new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 201 });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p1/storage")) {
        return new Response(JSON.stringify({ db_path: "C:/Cabinet/profiles/p1/cabinet.db", media_dir: "C:/Cabinet/profiles/p1/media" }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");

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
    localStorage.setItem("cabinet.workspace.p1", "0");

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

  it("auto-loads onboarding sample data after registration finish", async () => {
    let sampleSeeded = false;
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
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/items") {
        if (sampleSeeded) {
          return new Response(JSON.stringify({ items: [{ id: "seed-1", part_number: "CAB-DEMO-001", title: "Starter Item" }] }), { status: 200 });
        }
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      if (url === "/api/auth/webauthn/register/begin" && init?.method === "POST") {
        return new Response(JSON.stringify({ session_id: "sess-reg-1", options: {} }), { status: 200 });
      }
      if (url === "/api/auth/webauthn/register/finish" && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url === "/api/onboarding/sample-data" && init?.method === "POST") {
        sampleSeeded = true;
        return new Response(
          JSON.stringify({ created_items: 1, created_wishlist_entries: 1, total_items: 1, total_wishlist_entries: 1, already_seeded_for_profile: false }),
          { status: 200 },
        );
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const begin = await screen.findByRole("button", { name: /begin webauthn registration/i });
    begin.click();
    const finish = await screen.findByRole("button", { name: /finish registration/i });
    finish.click();

    expect(await screen.findByText(/onboarding sample data loaded/i)).toBeInTheDocument();
    expect(await screen.findByText(/current items: 1/i)).toBeInTheDocument();
  });

  it("defaults to starter view and reveals advanced workspace on demand", async () => {
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
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    expect(await screen.findByRole("button", { name: /open advanced workspace/i })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /discovery scanner/i })).not.toBeInTheDocument();

    const openAdvanced = await screen.findByRole("button", { name: /open advanced workspace/i });
    openAdvanced.click();
    expect(await screen.findByRole("heading", { name: /discovery scanner/i })).toBeInTheDocument();
  });

  it("shows step-scoped starter onboarding actions", async () => {
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
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "2");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.path.p1", "quick");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    expect(await screen.findByRole("button", { name: /complete identity/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /load sample data/i })).not.toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /open advanced workspace/i })).toBeInTheDocument();
  });

  it("persists onboarding wizard step and resumes from last incomplete step", async () => {
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
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "1");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    expect(await screen.findByText(/step 1 of 5/i)).toBeInTheDocument();
    const next = await screen.findByRole("button", { name: /next step/i });
    next.click();
    expect(await screen.findByText(/step 2 of 5/i)).toBeInTheDocument();
    expect(localStorage.getItem("cabinet.onboarding.step.p1")).toBe("2");

    cleanup();
    render(<App />);
    const activateAgain = await screen.findByRole("button", { name: /use alpha/i });
    activateAgain.click();
    expect(await screen.findByText(/step 2 of 5/i)).toBeInTheDocument();
  });

  it("supports step-1 welcome setup choices and persists selected path", async () => {
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
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/onboarding/sample-data" && init?.method === "POST") {
        return new Response(JSON.stringify({ created_items: 1 }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "1");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");

    render(<App />);
    let activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    expect(await screen.findByRole("button", { name: /start setup/i })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /import existing collection/i })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /use sample data/i })).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /start setup/i }));
    expect(await screen.findByText(/step 2 of 5/i)).toBeInTheDocument();
    expect(localStorage.getItem("cabinet.onboarding.path.p1")).toBe("quick");

    cleanup();
    localStorage.setItem("cabinet.onboarding.step.p1", "1");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    render(<App />);
    activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    fireEvent.click(await screen.findByRole("button", { name: /import existing collection/i }));
    expect(await screen.findByText(/step 2 of 5/i)).toBeInTheDocument();
    expect(localStorage.getItem("cabinet.onboarding.path.p1")).toBe("import");

    cleanup();
    localStorage.setItem("cabinet.onboarding.step.p1", "1");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    render(<App />);
    activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    fireEvent.click(await screen.findByRole("button", { name: /use sample data/i }));
    expect(localStorage.getItem("cabinet.onboarding.path.p1")).toBe("sample");
    expect(await screen.findByText(/onboarding sample data loaded/i)).toBeInTheDocument();
  });

  it("shows step-3 starter data choices and persists start-empty selection", async () => {
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
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "3");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.path.p1", "quick");
    localStorage.setItem("cabinet.onboarding.identity_completed.p1", "1");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const sampleButton = await screen.findByRole("button", { name: /load sample data \(recommended\)/i });
    expect(sampleButton).toBeInTheDocument();
    const startEmpty = await screen.findByRole("button", { name: /start empty/i });
    fireEvent.click(startEmpty);

    expect(localStorage.getItem("cabinet.onboarding.starter_data.p1")).toBe("empty");
    expect(await screen.findByText(/starting with an empty collection/i)).toBeInTheDocument();
    expect(screen.queryByText(/onboarding sample data loaded/i)).not.toBeInTheDocument();
  });

  it("runs sample-data seeding from step 3 and handles idempotent reruns", async () => {
    let sampleCalls = 0;
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
      if (url === "/api/onboarding/sample-data" && init?.method === "POST") {
        sampleCalls += 1;
        if (sampleCalls === 1) {
          return new Response(JSON.stringify({ created_items: 3 }), { status: 200 });
        }
        return new Response(JSON.stringify({ created_items: 0 }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "3");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.path.p1", "quick");
    localStorage.setItem("cabinet.onboarding.identity_completed.p1", "1");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const sampleButton = await screen.findByRole("button", { name: /load sample data \(recommended\)/i });
    fireEvent.click(sampleButton);
    expect(await screen.findByText(/onboarding sample data loaded \(3 starter items\)/i)).toBeInTheDocument();
    expect(localStorage.getItem("cabinet.onboarding.starter_data.p1")).toBe("sample");

    fireEvent.click(sampleButton);
    expect(await screen.findByText(/onboarding sample data already available/i)).toBeInTheDocument();
    expect(sampleCalls).toBe(2);
  });

  it("handles step-4 quick add validation and advances to preferences after success", async () => {
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
      if (url === "/api/items" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      if (url === "/api/items" && init?.method === "POST") {
        return new Response(JSON.stringify({ id: "i1", part_number: "PN-900", title: "Wizard Item", brand: "AFX" }), { status: 201 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "4");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.path.p1", "quick");
    localStorage.setItem("cabinet.onboarding.identity_completed.p1", "1");
    localStorage.setItem("cabinet.onboarding.starter_data.p1", "empty");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    expect(await screen.findByText(/step 4 of 5/i)).toBeInTheDocument();
    expect(await screen.findByText(/current items: 0/i)).toBeInTheDocument();
    const advanced = screen.getByText(/advanced fields \(optional\)/i).closest("details");
    expect(advanced).not.toHaveAttribute("open");

    fireEvent.click(await screen.findByRole("button", { name: /add first item/i }));
    expect(await screen.findByText(/part number is required/i)).toBeInTheDocument();
    expect(await screen.findByText(/title is required/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/part number/i), { target: { value: "PN-900" } });
    fireEvent.change(screen.getByLabelText(/item title/i), { target: { value: "Wizard Item" } });
    fireEvent.change(screen.getByLabelText(/^brand$/i), { target: { value: "AFX" } });
    fireEvent.click(screen.getByRole("button", { name: /add first item/i }));

    expect(await screen.findByText(/first item added\. continue to preferences\./i)).toBeInTheDocument();
    expect(await screen.findByText(/step 5 of 5/i)).toBeInTheDocument();
    expect(await screen.findByText(/current items: 1/i)).toBeInTheDocument();
  });

  it("requires successful identity completion before progressing past step 2", async () => {
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
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/auth/webauthn/register/begin" && init?.method === "POST") {
        return new Response(JSON.stringify({ session_id: "sess-reg-1", options: {} }), { status: 200 });
      }
      if (url === "/api/auth/webauthn/register/finish" && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url === "/api/onboarding/sample-data" && init?.method === "POST") {
        return new Response(JSON.stringify({ created_items: 0 }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "2");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.path.p1", "quick");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const next = await screen.findByRole("button", { name: /next step/i });
    expect(next).toBeDisabled();

    const completeIdentity = await screen.findByRole("button", { name: /complete identity/i });
    completeIdentity.click();
    expect(await screen.findByText(/auth status: registration_finished/i)).toBeInTheDocument();
    expect(next).toBeEnabled();

    fireEvent.click(next);
    expect(await screen.findByText(/step 3 of 5/i)).toBeInTheDocument();
  });

  it("keeps step 2 blocked when identity action fails and allows retry", async () => {
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
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/auth/webauthn/register/begin" && init?.method === "POST") {
        return new Response(JSON.stringify({ error: "failed" }), { status: 500 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "2");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.path.p1", "quick");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const next = await screen.findByRole("button", { name: /next step/i });
    expect(next).toBeDisabled();

    const completeIdentity = await screen.findByRole("button", { name: /complete identity/i });
    completeIdentity.click();
    expect(await screen.findByText(/request_failed:\/api\/auth\/webauthn\/register\/begin/i)).toBeInTheDocument();
    expect(next).toBeDisabled();
    expect(completeIdentity).toBeInTheDocument();
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
    expect((await screen.findAllByText(/pn-001/i)).length).toBeGreaterThan(0);

    const partInput = await screen.findByLabelText(/part number/i);
    const titleInput = await screen.findByLabelText(/item title/i);
    const brandInput = await screen.findByLabelText(/^brand$/i);
    const addButton = await screen.findByRole("button", { name: /add item/i });

    fireEvent.change(partInput, { target: { value: "PN-002" } });
    fireEvent.change(titleInput, { target: { value: "New Item" } });
    fireEvent.change(brandInput, { target: { value: "Hot Wheels" } });
    addButton.click();

    expect((await screen.findAllByText(/pn-002/i)).length).toBeGreaterThan(0);
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

  it("opens fullscreen photo preview and handles camera permission errors", async () => {
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
    vi.stubGlobal("navigator", {
      mediaDevices: {
        getUserMedia: vi.fn().mockRejectedValue(new Error("denied")),
      },
    });

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const loadPhotos = await screen.findByRole("button", { name: /load photos/i });
    loadPhotos.click();
    const openFullscreen = await screen.findByRole("button", { name: /open fullscreen preview/i });
    openFullscreen.click();
    expect(await screen.findByText(/fullscreen: a.jpg/i)).toBeInTheDocument();

    const openCamera = await screen.findByRole("button", { name: /open camera/i });
    openCamera.click();
    expect(await screen.findByText(/camera_unavailable/i)).toBeInTheDocument();
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
      if (url.includes("/api/pricing/history/export?item_id=")) {
        return new Response("date,price\n2026-02-21,18", { status: 200 });
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
    const exportPricing = await screen.findByRole("button", { name: /export pricing history/i });
    exportPricing.click();
    expect(await screen.findByText(/export bytes: 24/i)).toBeInTheDocument();
  });

  it("loads and saves profile settings and secrets", async () => {
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
      if (url.includes("/api/profiles/p1/settings") && (!init || init.method === undefined)) {
        return new Response(
          JSON.stringify({ settings: { scanner_schedule: "0 6 * * *", backup_frequency: "daily", "storage.db_path": "/tmp/p1.db" } }),
          { status: 200 },
        );
      }
      if (url.includes("/api/profiles/p1/settings") && init?.method === "PUT") {
        return new Response(
          JSON.stringify({ settings: { scanner_schedule: "0 12 * * *", backup_frequency: "hourly", "storage.db_path": "/tmp/new.db" } }),
          { status: 200 },
        );
      }
      if (url.includes("/api/profiles/p1/secrets") && init?.method === "PUT") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url.includes("/api/settings/reset-ignore-rules") && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url.includes("/api/items/i1/photos-rebuild") && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const loadProfileSettings = await screen.findByRole("button", { name: /load profile settings/i });
    loadProfileSettings.click();
    expect(await screen.findByDisplayValue(/0 6 \* \* \*/i)).toBeInTheDocument();

    fireEvent.change(await screen.findByLabelText(/scanner schedule/i), { target: { value: "0 12 * * *" } });
    fireEvent.change(await screen.findByLabelText(/backup frequency/i), { target: { value: "hourly" } });
    fireEvent.change(await screen.findByLabelText(/database path/i), { target: { value: "/tmp/new.db" } });
    const saveProfileSettings = await screen.findByRole("button", { name: /save profile settings/i });
    saveProfileSettings.click();
    expect(await screen.findByText(/settings_saved/i)).toBeInTheDocument();

    fireEvent.change(await screen.findByLabelText(/openai api key/i), { target: { value: "sk-test" } });
    fireEvent.change(await screen.findByLabelText(/ebay app id/i), { target: { value: "app-id" } });
    fireEvent.change(await screen.findByLabelText(/ebay auth token/i), { target: { value: "token" } });
    const saveSecrets = await screen.findByRole("button", { name: /save secrets/i });
    saveSecrets.click();
    expect(await screen.findByText(/secrets_saved/i)).toBeInTheDocument();

    const resetIgnore = await screen.findByRole("button", { name: /reset ignore rules/i });
    resetIgnore.click();
    expect(await screen.findByText(/ignore_rules_reset_ok/i)).toBeInTheDocument();
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

  it("toggles AI and applies suggestion with explicit confirmation gate", async () => {
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
      if (url === "/api/ai/toggle" && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url === "/api/ai/suggest/title" && init?.method === "POST") {
        return new Response(JSON.stringify({ title: "Suggested Title", confidence: 0.92 }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    localStorage.setItem("cabinet.workspace.p1", "1");
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const openAdvanced = await screen.findByRole("button", { name: /open advanced workspace/i });
    openAdvanced.click();
    const toggleAi = await screen.findByRole("button", { name: /enable ai/i });
    toggleAi.click();
    const titleInput = await screen.findByLabelText(/title normalization input/i);
    fireEvent.change(titleInput, { target: { value: "AFX P-1 listing" } });
    const suggest = await screen.findByRole("button", { name: /normalize title/i });
    suggest.click();
    expect(await screen.findByText(/ai confidence: 0.92/i)).toBeInTheDocument();
    const apply = await screen.findByRole("button", { name: /apply suggestion/i });
    expect(apply).toBeDisabled();
    const confirmApply = await screen.findByLabelText(/confirm apply suggestion/i);
    fireEvent.click(confirmApply);
    expect(apply).toBeEnabled();
    apply.click();
    expect(await screen.findByDisplayValue(/suggested title/i)).toBeInTheDocument();
  });

  it("supports debounced collection search and saved filters", async () => {
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
      if (url.startsWith("/api/search/items?")) {
        return new Response(JSON.stringify({ items: [{ id: "i777", part_number: "PN-777", title: "AFX Search Hit", brand: "AFX" }] }), {
          status: 200,
        });
      }
      if (url.includes("/api/profiles/p1/saved-filters") && (!init || init.method === undefined)) {
        return new Response(
          JSON.stringify({ saved_filters: [{ id: "f1", name: "AFX Only", query: { text: "AFX", brand: "AFX", sort_by: "part_number" } }] }),
          { status: 200 },
        );
      }
      if (url.includes("/api/profiles/p1/saved-filters") && init?.method === "POST") {
        return new Response(
          JSON.stringify({ id: "f2", name: "My Filter", query: { text: "AFX", brand: "AFX", sort_by: "part_number" } }),
          { status: 201 },
        );
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    fireEvent.change(await screen.findByLabelText(/collection search/i), { target: { value: "AFX" } });
    fireEvent.change(await screen.findByLabelText(/collection brand filter/i), { target: { value: "AFX" } });
    fireEvent.change(await screen.findByLabelText(/collection sort/i), { target: { value: "part_number" } });
    await new Promise((resolve) => setTimeout(resolve, 350));

    expect((await screen.findAllByText(/pn-777/i)).length).toBeGreaterThan(0);

    const loadSaved = await screen.findByRole("button", { name: /load saved filters/i });
    loadSaved.click();
    expect(await screen.findByText(/afx only/i)).toBeInTheDocument();

    fireEvent.change(await screen.findByLabelText(/saved filter name/i), { target: { value: "My Filter" } });
    const saveFilter = await screen.findByRole("button", { name: /save current filter/i });
    saveFilter.click();
    expect(await screen.findByText(/my filter/i)).toBeInTheDocument();

  });

  it("loads matching results and not-in-collection panel", async () => {
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
      if (url === "/api/matching/results") {
        return new Response(
          JSON.stringify({
            results: [
              { candidate_id: "c1", state: "matched", part_number: "PN-1" },
              { candidate_id: "c2", state: "suggested", part_number: "PN-2" },
              { candidate_id: "c3", state: "not_in_collection", part_number: "PN-3" },
            ],
          }),
          { status: 200 },
        );
      }
      if (url.startsWith("/api/discovery/not-in-collection?")) {
        return new Response(
          JSON.stringify({
            items: [{ candidate_id: "c3", title: "AFX P-9", price: 12, url: "http://example.local/listing/c3", last_seen: "2026-02-21" }],
          }),
          { status: 200 },
        );
      }
      if (url === "/api/discovery/action" && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const loadMatching = await screen.findByRole("button", { name: /load matching results/i });
    loadMatching.click();
    expect(await screen.findByText(/matched: 1/i)).toBeInTheDocument();
    expect(await screen.findByText(/suggested: 1/i)).toBeInTheDocument();
    expect(await screen.findByText(/not in collection: 1/i)).toBeInTheDocument();

    fireEvent.change(await screen.findByLabelText(/not in collection query/i), { target: { value: "AFX" } });
    fireEvent.change(await screen.findByLabelText(/not in collection max price/i), { target: { value: "20" } });
    fireEvent.change(await screen.findByLabelText(/not in collection date from/i), { target: { value: "2026-02-01" } });

    const loadNotOwned = await screen.findByRole("button", { name: /load not in collection/i });
    loadNotOwned.click();
    expect(await screen.findByText(/afx p-9/i)).toBeInTheDocument();

    const createItem = await screen.findByRole("button", { name: /create item/i });
    createItem.click();
  });

  it("loads pricing source breakdown and wishlist below-target indicator", async () => {
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
      if (url === "/api/wishlist") {
        return new Response(
          JSON.stringify({ wishlist: [{ id: "w1", item_id: "i1", target_price: 25, below_target_now: true, priority: "high" }] }),
          { status: 200 },
        );
      }
      if (url.includes("/api/pricing/by-source?item_id=i1")) {
        return new Response(
          JSON.stringify({
            by_source: {
              ebay: [{ snapshot_date: "2026-02-21", min_price: 10, median_price: 11, latest_price: 12, source: "ebay" }],
            },
          }),
          { status: 200 },
        );
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    const loadWishlist = await screen.findByRole("button", { name: /load wishlist/i });
    loadWishlist.click();
    expect(await screen.findByText(/below target/i)).toBeInTheDocument();

    const loadSources = await screen.findByRole("button", { name: /load pricing sources/i });
    loadSources.click();
    expect(await screen.findByText(/source groups: 1/i)).toBeInTheDocument();
    expect(await screen.findByText(/ebay: 1 snapshots/i)).toBeInTheDocument();
  });
});
