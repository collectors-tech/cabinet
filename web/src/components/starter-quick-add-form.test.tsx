import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { StarterQuickAddForm, type StarterQuickAddValues } from "./starter-quick-add-form";

describe("StarterQuickAddForm", () => {
  afterEach(() => {
    cleanup();
  });

  it("shows required fields and keeps advanced fields collapsed by default", () => {
    const onSubmit = vi.fn();
    render(<StarterQuickAddForm onSubmit={onSubmit} />);

    expect(screen.getByLabelText(/part number/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/item title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^brand$/i)).toBeInTheDocument();
    const advanced = screen.getByText(/advanced fields \(optional\)/i).closest("details");
    expect(advanced).not.toBeNull();
    expect(advanced).not.toHaveAttribute("open");
  });

  it("shows inline validation for invalid submit and allows valid submit", async () => {
    const onSubmit = vi.fn(async (_values: StarterQuickAddValues) => {});
    render(<StarterQuickAddForm onSubmit={onSubmit} />);

    fireEvent.click(screen.getByRole("button", { name: /add first item/i }));
    expect(await screen.findByText(/part number is required/i)).toBeInTheDocument();
    expect(await screen.findByText(/title is required/i)).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();

    fireEvent.change(screen.getByLabelText(/part number/i), { target: { value: "PN-700" } });
    fireEvent.change(screen.getByLabelText(/item title/i), { target: { value: "Starter Car" } });
    fireEvent.change(screen.getByLabelText(/^brand$/i), { target: { value: "AFX" } });
    fireEvent.click(screen.getByRole("button", { name: /add first item/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          part_number: "PN-700",
          title: "Starter Car",
          brand: "AFX",
        }),
      );
    });
  });
});
