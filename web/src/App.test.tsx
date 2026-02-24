import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
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

  it("uses fixed shell semantics with sticky topbar and scrolling content container", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    expect(screen.getByRole("heading", { name: /welcome to cabinet/i })).toBeInTheDocument();
    expect(document.querySelector(".cabinet-content-scroll")).toBeTruthy();
    expect(document.querySelector(".cabinet-topbar")).toBeTruthy();
    expect(document.querySelector(".cabinet-sidebar")).toBeTruthy();
    expect(screen.getByLabelText(/collection context pane/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/primary content/i)).toBeInTheDocument();
  });

  it("exposes semantic shell classes for page-header and primary-nav", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    expect(document.querySelector("header.page-header")).toBeTruthy();
    expect(document.querySelector("aside.primary-nav")).toBeTruthy();
    expect(document.querySelector("aside.collection-context-pane")).toBeTruthy();
  });

  it("updates active context label from the context pane", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /wishlist focus/i }));
    expect(screen.getByText(/context: wishlist focus/i)).toBeInTheDocument();
  });

  it("collapses and expands the collection context pane", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    const collapse = screen.getByRole("button", { name: /collapse collection pane/i });
    fireEvent.click(collapse);
    expect(screen.getByTestId("app-shell")).toHaveClass("cabinet-shell-context-collapsed");

    const expand = screen.getByRole("button", { name: /expand collection pane/i });
    fireEvent.click(expand);
    expect(screen.getByTestId("app-shell")).not.toHaveClass("cabinet-shell-context-collapsed");
  });

  it("keeps diagnostics links collapsed behind diagnostics summary", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);
    const diagnostics = document.querySelector(".cabinet-home-diagnostics");
    expect(diagnostics).toBeTruthy();
    expect(diagnostics).not.toHaveAttribute("open");
    fireEvent.click(screen.getByText(/^diagnostics$/i));
    expect(diagnostics).toHaveAttribute("open");
    const link = await screen.findByRole("link", { name: /api kitchen sync/i });
    expect(link).toHaveAttribute("href", "/apidocs");
  });

  it("opens and closes the global chat copilot rail", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    const openChat = await screen.findByRole("button", { name: /toggle chat copilot/i });
    fireEvent.click(openChat);
    expect(await screen.findByRole("complementary", { name: /chat copilot/i })).toBeInTheDocument();

    const closeChat = await screen.findByRole("button", { name: /close chat copilot/i });
    fireEvent.click(closeChat);
    expect(screen.queryByRole("complementary", { name: /chat copilot/i })).not.toBeInTheDocument();
  });

  it("persists chat rail open state in local storage", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    localStorage.setItem("cabinet.chat.open", "1");
    render(<App />);
    expect(await screen.findByRole("complementary", { name: /chat copilot/i })).toBeInTheDocument();
  });

  it("supports context chips that prefill the chat composer", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    fireEvent.click(await screen.findByRole("button", { name: /toggle chat copilot/i }));
    fireEvent.click(await screen.findByRole("button", { name: /wishlist hits context chip/i }));

    expect((screen.getByLabelText(/chat message/i) as HTMLTextAreaElement).value).toMatch(/wishlist hits/i);
  });

  it("uses preview and confirm flow before applying chat suggested actions", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Default" }] }), { status: 200 });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p1/storage")) {
        return new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 });
      }
      if (url.includes("/api/auth/requirements?profile_id=p1")) {
        return new Response(JSON.stringify({ requires_registration: false }), { status: 200 });
      }
      if (url.includes("/api/dashboard")) {
        return new Response(JSON.stringify({ new_discoveries: 2, wishlist_hits: 1, price_drops: 1, recently_added: 3, total_items: 10, total_instances: 14 }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use default/i }));
    fireEvent.click(await screen.findByRole("button", { name: /toggle chat copilot/i }));

    fireEvent.click(await screen.findByRole("button", { name: /preview open collection workspace action/i }));
    expect(await screen.findByText(/ready to apply: open collection workspace/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /confirm apply action/i }));
    expect(await screen.findByRole("heading", { name: /^collection$/i })).toBeInTheDocument();
  });

  it("sends a chat message and appends an assistant response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    fireEvent.click(await screen.findByRole("button", { name: /toggle chat copilot/i }));
    fireEvent.change(screen.getByLabelText(/chat message/i), { target: { value: "Show collection gaps" } });
    fireEvent.click(screen.getByRole("button", { name: /^send$/i }));

    expect(await screen.findByText(/show collection gaps/i)).toBeInTheDocument();
    expect(await screen.findByText(/base chat scaffold is active/i)).toBeInTheDocument();
  });

  it("supports adding and removing local attachments before sending", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    fireEvent.click(await screen.findByRole("button", { name: /toggle chat copilot/i }));
    const input = await screen.findByLabelText(/chat attachment/i);
    const file = new File(["sample"], "notes.txt", { type: "text/plain" });
    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByText(/notes\.txt/i)).toBeInTheDocument();
    fireEvent.click(await screen.findByRole("button", { name: /remove attachment notes\.txt/i }));
    expect(screen.queryByText(/notes\.txt/i)).not.toBeInTheDocument();
  });

  it("shows app version and build date in left nav footer", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/api/profiles") {
        return new Response(JSON.stringify({ profiles: [] }), { status: 200 });
      }
      if (url === "/api/runtime") {
        return new Response(
          JSON.stringify({
            update_channel: "stable",
            update_public_key_configured: false,
            app_version: "rev-abc1234",
            build_date: "2026-02-24T03:00:00Z",
          }),
          { status: 200 },
        );
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const meta = await screen.findByLabelText(/app build metadata/i);
    expect(meta).toHaveTextContent(/version:\s*rev-abc1234/i);
    expect(meta).toHaveTextContent(/build date:\s*2026-02-24t03:00:00z/i);
  });

  it("opens and closes mobile navigation drawer", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);
    const openButton = await screen.findByRole("button", { name: /open navigation menu/i });
    fireEvent.click(openButton);
    const drawer = await screen.findByRole("dialog", { name: /navigation menu/i });
    expect(drawer).toBeInTheDocument();
    expect(within(drawer).getByLabelText(/collection context pane/i)).toBeInTheDocument();

    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("dialog", { name: /navigation menu/i })).not.toBeInTheDocument();
  });

  it("collapses primary nav to icon mode and restores expanded mode", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);

    const collapse = await screen.findByRole("button", { name: /collapse primary navigation/i });
    fireEvent.click(collapse);
    expect(document.querySelector(".cabinet-sidebar-collapsed")).toBeTruthy();
    expect(screen.getByRole("button", { name: /^dashboard$/i })).toHaveAttribute("title", "Dashboard");

    const expand = await screen.findByRole("button", { name: /expand primary navigation/i });
    fireEvent.click(expand);
    expect(document.querySelector(".cabinet-sidebar-collapsed")).toBeFalsy();
  });

  it("supports nav main reordering and visibility toggles from nav editor", async () => {
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
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));

    const navMain = screen.getAllByRole("navigation")[0];
    expect(within(navMain).getAllByRole("button")[0]).toHaveAccessibleName("Dashboard");

    fireEvent.click(screen.getByRole("button", { name: /edit nav main/i }));
    expect(screen.getByLabelText(/nav main editor/i)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^dashboard$/i })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /move dashboard down/i }));
    fireEvent.click(screen.getByRole("button", { name: /finish nav main editing/i }));

    const navMainAfterMove = screen.getAllByRole("navigation")[0];
    expect(within(navMainAfterMove).getAllByRole("button")[0]).toHaveAccessibleName("Collection");

    fireEvent.click(screen.getByRole("button", { name: /edit nav main/i }));
    fireEvent.click(screen.getByRole("button", { name: /hide collection/i }));
    fireEvent.click(screen.getByRole("button", { name: /finish nav main editing/i }));

    const navMainAfterHide = screen.getAllByRole("navigation")[0];
    expect(within(navMainAfterHide).queryByRole("button", { name: /^collection$/i })).not.toBeInTheDocument();
  });

  it("renders semantic collection browser base layout in advanced workspace", async () => {
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
        return new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-1", title: "AFX Camaro", brand: "AFX", category: "Cars" }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^collection$/i }));

    expect(await screen.findByTestId("collection-browser")).toBeInTheDocument();
    expect(screen.getByTestId("collection-tree")).toBeInTheDocument();
    expect(screen.getByTestId("collection-results")).toBeInTheDocument();
    expect(screen.getByTestId("collection-summary-strip")).toBeInTheDocument();
  });

  it("renders eye icon toggle control for nav item visibility in edit mode", async () => {
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
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(screen.getByRole("button", { name: /edit nav main/i }));

    const hideCollection = screen.getByRole("button", { name: /hide collection/i });
    expect(hideCollection.querySelector("svg")).toBeTruthy();
  });

  it("switches visible advanced workspace section from sidebar navigation", async () => {
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
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));

    const dashboardNav = await screen.findByRole("button", { name: /^dashboard$/i });
    fireEvent.click(dashboardNav);
    expect(dashboardNav).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("heading", { name: /^dashboard$/i })).toBeInTheDocument();

    const collectionNav = await screen.findByRole("button", { name: /^collection$/i });
    fireEvent.click(collectionNav);
    expect(collectionNav).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("heading", { name: /^collection$/i })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /^dashboard$/i })).not.toBeInTheDocument();

    const discoveriesNav = await screen.findByRole("button", { name: /^discoveries$/i });
    fireEvent.click(discoveriesNav);
    expect(discoveriesNav).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("heading", { name: /^not in my collection$/i })).toBeInTheDocument();

    const aiAssistNav = await screen.findByRole("button", { name: /^ai assist$/i });
    fireEvent.click(aiAssistNav);
    expect(aiAssistNav).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("heading", { name: /^ai assist$/i })).toBeInTheDocument();

    const barcodesNav = await screen.findByRole("button", { name: /^barcodes$/i });
    fireEvent.click(barcodesNav);
    expect(barcodesNav).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("heading", { name: /^barcodes$/i })).toBeInTheDocument();

    const photosNav = await screen.findByRole("button", { name: /^photos$/i });
    fireEvent.click(photosNav);
    expect(photosNav).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("heading", { name: /^photos$/i })).toBeInTheDocument();

    const settingsNav = await screen.findByRole("button", { name: /^settings$/i });
    fireEvent.click(settingsNav);
    expect(settingsNav).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("heading", { name: /settings and diagnostics/i })).toBeInTheDocument();
  });

  it("uses mobile drawer navigation to switch active screen", async () => {
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
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));

    fireEvent.click(await screen.findByRole("button", { name: /open navigation menu/i }));
    const drawer = await screen.findByRole("dialog", { name: /navigation menu/i });
    const scannerNavInDrawer = within(drawer).getByRole("button", { name: /^scanner$/i });
    fireEvent.click(scannerNavInDrawer);

    expect(screen.queryByRole("dialog", { name: /navigation menu/i })).not.toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: /discovery scanner/i })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /^scanner$/i })).toHaveAttribute("aria-current", "page");
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

  it("minimizes onboarding presentation once advanced workspace is active", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Default" }] }), { status: 200 });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p1/storage")) {
        return new Response(JSON.stringify({ db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" }), { status: 200 });
      }
      if (url.includes("/api/auth/requirements?profile_id=p1")) {
        return new Response(JSON.stringify({ requires_registration: false }), { status: 200 });
      }
      if (url.includes("/api/dashboard")) {
        return new Response(
          JSON.stringify({
            new_discoveries: 2,
            wishlist_hits: 1,
            price_drops: 1,
            recently_added: 3,
            total_items: 10,
            total_instances: 14,
            estimated_value: 1000,
          }),
          { status: 200 },
        );
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-1", title: "AFX Camaro", brand: "AFX", category: "Cars" }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use default/i }));

    expect(screen.queryByRole("heading", { name: /starter onboarding wizard/i })).not.toBeInTheDocument();
    expect(await screen.findByText(/onboarding complete/i)).toBeInTheDocument();
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
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === "/api/profiles" && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ profiles: [{ id: "p2", name: "Beta" }] }), { status: 200 });
      }
      if (url === "/api/profiles/active" && init?.method === "PUT") {
        return new Response(JSON.stringify({ id: "p2", name: "Beta" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p2/storage")) {
        return new Response(JSON.stringify({ db_path: "/tmp/p2.db", media_dir: "/tmp/p2/media" }), { status: 200 });
      }
      if (url.includes("/api/auth/requirements?profile_id=p2")) {
        return new Response(JSON.stringify({ requires_registration: true }), { status: 200 });
      }
      if (url === "/api/items") {
        return new Response(JSON.stringify({ items: [] }), { status: 200 });
      }
      if (url === "/api/auth/webauthn/register/begin" && init?.method === "POST") {
        return new Response(JSON.stringify({ session_id: "sess-reg-1", options: {} }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
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

  it("shows advanced workspace controls only after onboarding completion", async () => {
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
    localStorage.setItem("cabinet.onboarding.completed.p1", "1");
    localStorage.setItem("cabinet.onboarding.step.p1", "5");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    expect(await screen.findByRole("button", { name: /open advanced workspace/i })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /discovery scanner/i })).not.toBeInTheDocument();

    const openAdvanced = await screen.findByRole("button", { name: /open advanced workspace/i });
    openAdvanced.click();
    expect(await screen.findByRole("heading", { name: /discovery scanner/i })).toBeInTheDocument();
  });

  it("keeps advanced workspace hidden pre-completion even when workspace flag is set", async () => {
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
    localStorage.setItem("cabinet.workspace.p1", "1");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "2");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    expect(await screen.findByRole("heading", { name: /starter onboarding wizard/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /open advanced workspace/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /discovery scanner/i })).not.toBeInTheDocument();
    expect(localStorage.getItem("cabinet.workspace.p1")).toBe("0");
  });

  it("supports legacy migration path by honoring workspace flag when onboarding state is absent", async () => {
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
    localStorage.setItem("cabinet.workspace.p1", "1");
    localStorage.removeItem("cabinet.onboarding.completed.p1");
    localStorage.removeItem("cabinet.onboarding.step.p1");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

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
    expect(screen.queryByRole("button", { name: /open advanced workspace/i })).not.toBeInTheDocument();
  });

  it("renders structured onboarding layout with progress rail and status panel", async () => {
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
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));

    expect(await screen.findByRole("heading", { name: /getting started/i })).toBeInTheDocument();
    expect(await screen.findByRole("list", { name: /onboarding progress/i })).toBeInTheDocument();
    expect(await screen.findByText(/quick status/i)).toBeInTheDocument();
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

  it("persists step-5 preferences, completes onboarding, and reopens in advanced workspace", async () => {
    let settingsPayload: Record<string, unknown> | null = null;
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
      if (url === "/api/profiles/p1/settings" && init?.method === "PUT") {
        settingsPayload = JSON.parse(String(init.body || "{}")) as Record<string, unknown>;
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "0");
    localStorage.setItem("cabinet.onboarding.step.p1", "5");
    localStorage.setItem("cabinet.onboarding.completed.p1", "0");
    localStorage.setItem("cabinet.onboarding.path.p1", "quick");
    localStorage.setItem("cabinet.onboarding.identity_completed.p1", "1");
    localStorage.setItem("cabinet.onboarding.starter_data.p1", "empty");

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();

    fireEvent.change(await screen.findByLabelText(/onboarding theme/i), { target: { value: "dark" } });
    fireEvent.change(await screen.findByLabelText(/onboarding backup frequency/i), { target: { value: "weekly" } });
    fireEvent.change(await screen.findByLabelText(/onboarding scanner schedule/i), { target: { value: "weekly" } });
    fireEvent.click(await screen.findByRole("button", { name: /finish onboarding/i }));

    expect(await screen.findByText(/onboarding complete\. advanced workspace unlocked\./i)).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: /discovery scanner/i })).toBeInTheDocument();
    expect(localStorage.getItem("cabinet.onboarding.completed.p1")).toBe("1");
    expect(localStorage.getItem("cabinet.workspace.p1")).toBe("1");
    expect(settingsPayload).toEqual(
      expect.objectContaining({
        settings: expect.objectContaining({
          backup_frequency: "weekly",
          scanner_schedule: "0 8 * * 1",
        }),
      }),
    );

    cleanup();
    render(<App />);
    const activateAgain = await screen.findByRole("button", { name: /use alpha/i });
    activateAgain.click();
    expect(await screen.findByRole("heading", { name: /discovery scanner/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /open advanced workspace/i })).not.toBeInTheDocument();
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
    let items = [{ id: "i1", part_number: "PN-001", title: "Existing", brand: "AFX" }];
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
        return new Response(JSON.stringify({ items }), { status: 200 });
      }
      if (url === "/api/items" && init?.method === "POST") {
        const created = { id: "i2", part_number: "PN-002", title: "New Item", brand: "Hot Wheels" };
        items = [...items, created];
        return new Response(JSON.stringify(created), { status: 201 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
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
        return new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-001", title: "Existing", brand: "AFX" }] }), { status: 200 });
      }
      if (url.includes("/api/items/i1/photos")) {
        return new Response(JSON.stringify({ photos: [{ id: "ph1", item_id: "i1", filename: "a.jpg", is_primary: true }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const loadPhotos = await screen.findByRole("button", { name: /load photos/i });
    loadPhotos.click();
    expect(await screen.findByText(/a.jpg/i)).toBeInTheDocument();
  });

  it("opens fullscreen photo preview and handles camera permission errors", async () => {
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
        return new Response(JSON.stringify({ items: [{ id: "i1", part_number: "PN-001", title: "Existing", brand: "AFX" }] }), { status: 200 });
      }
      if (url.includes("/api/items/i1/photos")) {
        return new Response(JSON.stringify({ photos: [{ id: "ph1", item_id: "i1", filename: "a.jpg", is_primary: true }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
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
      if (url === "/api/scanner/query-sets") {
        return new Response(JSON.stringify({ query_sets: [{ id: "q1", name: "AFX Search" }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use alpha/i });
    activate.click();
    const loadQuerySets = await screen.findByRole("button", { name: /load query sets/i });
    loadQuerySets.click();
    expect(await screen.findByText(/afx search/i)).toBeInTheDocument();
  });

  it("runs scanner scheduled workflows, provider health, and matching", async () => {
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
      if (url === "/api/scanner/run/scheduled" && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url === "/api/scanner/failures") {
        return new Response(JSON.stringify({ failures: [{ id: "f1", query_set_id: "q1", reason: "rate_limited", attempts: 2 }] }), { status: 200 });
      }
      if (url === "/api/scanner/failures/retry" && init?.method === "POST") {
        return new Response(JSON.stringify({ retry_started: true, query_set_id: "q1" }), { status: 200 });
      }
      if (url.startsWith("/api/provider/health?provider=ebay")) {
        return new Response(JSON.stringify({ provider: "ebay", state: "healthy", healthy: true }), { status: 200 });
      }
      if (url === "/api/matching/run" && init?.method === "POST") {
        return new Response(JSON.stringify({ processed: 3 }), { status: 200 });
      }
      if (url === "/api/matching/results") {
        return new Response(JSON.stringify({ results: [{ candidate_id: "c1", state: "matched" }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^scanner$/i }));

    fireEvent.click(await screen.findByRole("button", { name: /run scheduled/i }));
    expect(await screen.findByText(/scheduled run: scheduled_scans_triggered/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /load scanner failures/i }));
    expect(await screen.findByText(/failure: q1 \/ rate_limited \/ attempts 2/i)).toBeInTheDocument();
    fireEvent.click(await screen.findByRole("button", { name: /retry failure q1/i }));
    expect(await screen.findByText(/scanner retry status: retry_started_for_q1/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /check provider health/i }));
    expect(await screen.findByText(/provider health: ebay \/ healthy/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /run matching/i }));
    expect(await screen.findByText(/matching run status: matching_run_ok:3/i)).toBeInTheDocument();
    expect(await screen.findByText(/matched: 1/i)).toBeInTheDocument();
  });

  it("auto-loads dedicated dashboard view and supports pricing actions", async () => {
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

    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));
    expect(await screen.findByRole("button", { name: /refresh dashboard/i })).toBeInTheDocument();
    expect(await screen.findByText(/new discoveries: 3/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^pricing$/i }));
    const loadWishlist = await screen.findByRole("button", { name: /^load wishlist$/i });
    loadWishlist.click();
    expect(await screen.findByText(/wishlist item: i1 target 25/i)).toBeInTheDocument();

    const loadGraph = await screen.findByRole("button", { name: /load pricing graph/i });
    loadGraph.click();
    expect(await screen.findByText(/pricing points: 2/i)).toBeInTheDocument();
  });

  it("shows dashboard empty and error states with refresh support", async () => {
    let dashboardCalls = 0;
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
      if (url === "/api/dashboard") {
        dashboardCalls += 1;
        if (dashboardCalls === 1) {
          return new Response(JSON.stringify({ error: "fail" }), { status: 500 });
        }
        return new Response(JSON.stringify({}), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));

    expect(await screen.findByText(/insight error: failed_to_load_dashboard/i)).toBeInTheDocument();
    fireEvent.click(await screen.findByRole("button", { name: /refresh dashboard/i }));
    expect(await screen.findByText(/new discoveries: 0/i)).toBeInTheDocument();
  });

  it("supports dashboard quick actions that deep-link into core workspaces", async () => {
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
        return new Response(JSON.stringify({ new_discoveries: 2, wishlist_hits: 1, price_drops: 1 }), { status: 200 });
      }
      if (url === "/api/wishlist") {
        return new Response(JSON.stringify({ wishlist: [{ id: "w1", item_id: "i1", target_price: 25 }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));

    fireEvent.click(await screen.findByRole("button", { name: /open pricing workspace/i }));
    expect(await screen.findByRole("button", { name: /^pricing$/i })).toHaveAttribute("aria-current", "page");
    expect(await screen.findByText(/wishlist item: i1 target 25/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));
    fireEvent.click(await screen.findByRole("button", { name: /open scanner workspace/i }));
    expect(await screen.findByRole("button", { name: /^scanner$/i })).toHaveAttribute("aria-current", "page");

    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));
    fireEvent.click(await screen.findByRole("button", { name: /open collection workspace/i }));
    expect(await screen.findByRole("button", { name: /^collection$/i })).toHaveAttribute("aria-current", "page");

    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));
    fireEvent.click(await screen.findByRole("button", { name: /open settings workspace/i }));
    expect(await screen.findByRole("button", { name: /^settings$/i })).toHaveAttribute("aria-current", "page");
  });

  it("renders home attention panel with actionable cards", async () => {
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
        return new Response(JSON.stringify({ new_discoveries: 5, wishlist_hits: 3, price_drops: 2, total_items: 10, total_instances: 15 }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));

    expect(await screen.findByRole("heading", { name: /what needs attention now/i })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /review discoveries/i })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /open pricing workspace/i })).toBeInTheDocument();
  });

  it("prioritizes action queue and shows next-best action guidance on dashboard", async () => {
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
        return new Response(JSON.stringify({ new_discoveries: 9, wishlist_hits: 2, price_drops: 1, recently_added: 4 }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));

    expect(await screen.findByRole("heading", { name: /action queue/i })).toBeInTheDocument();
    expect(await screen.findByText(/review new discoveries/i)).toBeInTheDocument();
    expect(await screen.findByText(/high priority/i)).toBeInTheDocument();
    expect(await screen.findByText(/next best action: review discoveries/i)).toBeInTheDocument();
  });

  it("shows calm-state guidance when no attention items are pending", async () => {
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
        return new Response(JSON.stringify({ new_discoveries: 0, wishlist_hits: 0, price_drops: 0 }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));

    expect(await screen.findByText(/everything looks stable today/i)).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /run scanner now/i })).toBeInTheDocument();
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
        return new Response(JSON.stringify({ logs: [{ event: "scanner_run_completed", created_at: "2026-02-23T00:00:00Z" }] }), { status: 200 });
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
    expect(await screen.findByText(/activity: scanner_run_completed/i)).toBeInTheDocument();
    const exportPricing = await screen.findByRole("button", { name: /export pricing history/i });
    exportPricing.click();
    expect(await screen.findByText(/export bytes: 24/i)).toBeInTheDocument();
  });

  it("supports backup list and guarded restore workflows", async () => {
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
      if (url === "/api/backup/list") {
        return new Response(JSON.stringify({ backups: ["/tmp/backups/cabinet-backup-20260223-120000.db", "/tmp/backups/fail-backup.db"] }), { status: 200 });
      }
      if (url === "/api/backup/restore" && init?.method === "POST") {
        const body = JSON.parse(String(init.body || "{}")) as { backup_path?: string };
        if (body.backup_path?.includes("fail")) {
          return new Response(JSON.stringify({ error: "failed_to_restore_backup" }), { status: 400 });
        }
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^settings$/i }));

    fireEvent.click(await screen.findByRole("button", { name: /load backups/i }));
    expect(await screen.findByText(/backup count: 2/i)).toBeInTheDocument();
    expect(await screen.findByRole("option", { name: /cabinet-backup-20260223-120000.db/i })).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /restore selected backup/i }));
    expect(await screen.findByText(/admin error: restore_confirmation_required/i)).toBeInTheDocument();

    fireEvent.change(await screen.findByLabelText(/backup selection/i), { target: { value: "/tmp/backups/fail-backup.db" } });
    fireEvent.click(await screen.findByLabelText(/confirm restore/i));
    fireEvent.click(await screen.findByRole("button", { name: /restore selected backup/i }));
    expect(await screen.findByText(/admin error: failed_to_restore_backup/i)).toBeInTheDocument();
    expect(await screen.findByText(/restore failed: verify the selected backup file is valid and readable/i)).toBeInTheDocument();

    fireEvent.change(await screen.findByLabelText(/backup selection/i), { target: { value: "/tmp/backups/cabinet-backup-20260223-120000.db" } });
    const confirmRestore = (await screen.findByLabelText(/confirm restore/i)) as HTMLInputElement;
    if (!confirmRestore.checked) {
      fireEvent.click(confirmRestore);
    }
    fireEvent.click(await screen.findByRole("button", { name: /restore selected backup/i }));
    expect(await screen.findByText(/settings status: backup_restored/i)).toBeInTheDocument();
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

  it("supports license import, status validation, and profile license sync", async () => {
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
      if (url.includes("/api/profiles/p1/license") && (!init || init.method === undefined)) {
        return new Response(JSON.stringify({ license_json: "{\"tier\":\"pro\"}" }), { status: 200 });
      }
      if (url.includes("/api/profiles/p1/license") && init?.method === "PUT") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url === "/api/license/import" && init?.method === "POST") {
        const payload = JSON.parse(String(init.body || "{}")) as { license?: { payload_base64?: string } };
        if (payload.license?.payload_base64 === "bad") {
          return new Response(JSON.stringify({ error: "failed_to_import_license" }), { status: 400 });
        }
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url.includes("/api/license/status?profile_id=p1")) {
        return new Response(
          JSON.stringify({ state: "valid", tier: "pro", features: ["ai_assist", "price_tracking"], expires_at: "2030-01-01T00:00:00Z" }),
          { status: 200 },
        );
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^settings$/i }));

    fireEvent.click(await screen.findByRole("button", { name: /load profile license/i }));
    expect(await screen.findByDisplayValue(/\{"tier":"pro"\}/i)).toBeInTheDocument();

    const payloadInput = await screen.findByLabelText(/license payload base64/i);
    fireEvent.change(payloadInput, { target: { value: "bad" } });
    fireEvent.change(await screen.findByLabelText(/license signature base64/i), { target: { value: "sig" } });
    fireEvent.click(await screen.findByRole("button", { name: /import license file/i }));
    expect(await screen.findByText(/admin error: failed_to_import_license/i)).toBeInTheDocument();

    fireEvent.change(payloadInput, { target: { value: "good" } });
    fireEvent.click(await screen.findByRole("button", { name: /import license file/i }));
    expect(await screen.findByText(/license import status: license_imported/i)).toBeInTheDocument();
    expect(await screen.findByText(/license validation: valid \/ pro/i)).toBeInTheDocument();
    expect(await screen.findByText(/license features: ai_assist, price_tracking/i)).toBeInTheDocument();
    expect(await screen.findByText(/license expires: 2030-01-01T00:00:00Z/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /save profile license/i }));
    expect(await screen.findByText(/settings status: license_profile_saved/i)).toBeInTheDocument();
  });

  it("supports diagnostics debug mode toggle and current state visibility", async () => {
    let debugEnabled = false;
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
        return new Response(JSON.stringify({ settings: { debug_mode: debugEnabled ? "true" : "false" } }), { status: 200 });
      }
      if (url === "/api/logs/debug" && init?.method === "POST") {
        const req = JSON.parse(String(init.body || "{}")) as { enabled?: boolean };
        debugEnabled = Boolean(req.enabled);
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^settings$/i }));

    fireEvent.click(await screen.findByRole("button", { name: /load profile settings/i }));
    expect(await screen.findByText(/debug mode: disabled/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /enable debug mode/i }));
    expect(await screen.findByText(/debug mode: enabled/i)).toBeInTheDocument();
    expect(await screen.findByText(/settings status: debug_mode_enabled/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /disable debug mode/i }));
    expect(await screen.findByText(/debug mode: disabled/i)).toBeInTheDocument();
    expect(await screen.findByText(/settings status: debug_mode_disabled/i)).toBeInTheDocument();
  });

  it("loads runtime and recovery diagnostics in settings", async () => {
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
      if (url === "/api/runtime") {
        return new Response(JSON.stringify({ update_channel: "stable", update_public_key_configured: true }), { status: 200 });
      }
      if (url === "/api/runtime/recovery") {
        return new Response(JSON.stringify({ recovery_required: false }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^settings$/i }));
    fireEvent.click(await screen.findByRole("button", { name: /load runtime diagnostics/i }));

    expect(await screen.findByText(/runtime channel: stable/i)).toBeInTheDocument();
    expect(await screen.findByText(/runtime signing key configured: yes/i)).toBeInTheDocument();
    expect(await screen.findByText(/recovery required: no/i)).toBeInTheDocument();
  });

  it("shows recovery-required banner when runtime recovery state is required", async () => {
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
      if (url === "/api/runtime/recovery") {
        return new Response(JSON.stringify({ recovery_required: true }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));

    expect(await screen.findByRole("alert", { name: /recovery required/i })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /open settings diagnostics/i })).toBeInTheDocument();
  });

  it("hides recovery-required banner when runtime recovery state is clear", async () => {
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
      if (url === "/api/runtime/recovery") {
        return new Response(JSON.stringify({ recovery_required: false }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));

    expect(screen.queryByRole("alert", { name: /recovery required/i })).not.toBeInTheDocument();
  });

  it("shows runtime diagnostics error when runtime endpoints fail", async () => {
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
      if (url === "/api/runtime") {
        return new Response(JSON.stringify({ message: "failed" }), { status: 500 });
      }
      if (url === "/api/runtime/recovery") {
        return new Response(JSON.stringify({ recovery_required: false }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^settings$/i }));
    fireEvent.click(await screen.findByRole("button", { name: /load runtime diagnostics/i }));

    expect(await screen.findByText(/admin error: failed_to_load_runtime_diagnostics/i)).toBeInTheDocument();
    expect(await screen.findByText(/runtime channel: unknown/i)).toBeInTheDocument();
  });

  it("mounts explicit top-level screen containers during nav transitions", async () => {
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
        return new Response(JSON.stringify({ new_discoveries: 1, wishlist_hits: 1, price_drops: 0, recently_added: 1, total_items: 1, total_instances: 1 }), {
          status: 200,
        });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));

    fireEvent.click(await screen.findByRole("button", { name: /^dashboard$/i }));
    expect(await screen.findByTestId("screen-dashboard")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^collection$/i }));
    expect(await screen.findByTestId("screen-collection")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^scanner$/i }));
    expect(await screen.findByTestId("screen-scanner")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^discoveries$/i }));
    expect(await screen.findByTestId("screen-discoveries")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^ai assist$/i }));
    expect(await screen.findByTestId("screen-ai")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^barcodes$/i }));
    expect(await screen.findByTestId("screen-barcodes")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^photos$/i }));
    expect(await screen.findByTestId("screen-photos")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^pricing$/i }));
    expect(await screen.findByTestId("screen-pricing")).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /^settings$/i }));
    expect(await screen.findByTestId("screen-settings")).toBeInTheDocument();
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
    const openAdvanced = screen.queryByRole("button", { name: /open advanced workspace/i });
    if (openAdvanced) {
      openAdvanced.click();
    }
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

    const loadWishlist = await screen.findByRole("button", { name: /^load wishlist$/i });
    loadWishlist.click();
    expect(await screen.findByText(/below target/i)).toBeInTheDocument();

    const loadSources = await screen.findByRole("button", { name: /load pricing sources/i });
    loadSources.click();
    expect(await screen.findByText(/source groups: 1/i)).toBeInTheDocument();
    expect(await screen.findByText(/ebay: 1 snapshots/i)).toBeInTheDocument();
  });

  it("supports pricing track/history/stats/trend/snapshot and wishlist hits workflows", async () => {
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
      if (url === "/api/pricing/track" && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url.includes("/api/pricing/history?item_id=i1")) {
        return new Response(JSON.stringify({ history: [{ day: "2026-02-21", min: 10, median: 11, latest: 12 }] }), { status: 200 });
      }
      if (url.includes("/api/pricing/stats?item_id=i1")) {
        return new Response(JSON.stringify({ min: 10, median: 11, latest: 12 }), { status: 200 });
      }
      if (url.includes("/api/pricing/trend?item_id=i1")) {
        return new Response(JSON.stringify({ trend: "down" }), { status: 200 });
      }
      if (url === "/api/pricing/snapshot/run" && init?.method === "POST") {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      if (url === "/api/wishlist/hits") {
        return new Response(JSON.stringify({ hits: [{ item_id: "i1", listing_id: "l1", title: "Hit Item", price: 18 }] }), { status: 200 });
      }
      return new Response(JSON.stringify({}), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("cabinet.workspace.p1", "1");

    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: /use alpha/i }));
    fireEvent.click(await screen.findByRole("button", { name: /^pricing$/i }));

    fireEvent.click(await screen.findByRole("button", { name: /track pricing/i }));
    expect(await screen.findByText(/pricing track status: pricing_track_enabled/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /load pricing history/i }));
    expect(await screen.findByText(/pricing history points: 1/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /load pricing stats/i }));
    expect(await screen.findByText(/pricing stats loaded: yes/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /load pricing trend/i }));
    expect(await screen.findByText(/pricing trend loaded: yes/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /run pricing snapshot/i }));
    expect(await screen.findByText(/snapshot status: pricing_snapshot_completed/i)).toBeInTheDocument();

    fireEvent.click(await screen.findByRole("button", { name: /load wishlist hits/i }));
    expect(await screen.findByText(/wishlist hit: i1 \/ hit item \/ 18/i)).toBeInTheDocument();
  });
});
