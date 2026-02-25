import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

describe("shell foundation controls", () => {
  afterEach(() => {
    cleanup();
    localStorage.clear();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders workspace selector, global command search, density control, and profile menu trigger", () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })));
    render(<App />);

    expect(screen.getByLabelText(/workspace db selector/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/global command search/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /open density controls/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /open profile menu/i })).toBeInTheDocument();
  });

  it("applies compact density class from topbar controls", () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })));
    render(<App />);

    fireEvent.click(screen.getByRole("button", { name: /open density controls/i }));
    fireEvent.click(screen.getByRole("button", { name: /set compact density/i }));

    expect(screen.getByTestId("app-shell")).toHaveClass("cabinet-density-compact");
  });
});
