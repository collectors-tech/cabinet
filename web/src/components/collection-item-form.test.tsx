import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { CollectionItemForm, type CollectionItemValues } from "./collection-item-form";

describe("CollectionItemForm", () => {
  afterEach(() => {
    cleanup();
  });

  it("shows required and optional fields with advanced metadata collapsed by default", () => {
    const onSubmit = vi.fn();
    render(<CollectionItemForm onSubmit={onSubmit} />);

    expect(screen.getByLabelText(/part number/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/item title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^brand$/i)).toBeInTheDocument();
    const advanced = screen.getByText(/advanced metadata/i).closest("details");
    expect(advanced).not.toBeNull();
    expect(advanced).not.toHaveAttribute("open");
  });

  it("submits create flow with required values", async () => {
    const onSubmit = vi.fn(async (_values: CollectionItemValues) => {});
    render(<CollectionItemForm onSubmit={onSubmit} />);

    fireEvent.change(screen.getByLabelText(/part number/i), { target: { value: "PN-100" } });
    fireEvent.change(screen.getByLabelText(/item title/i), { target: { value: "Mainline" } });
    fireEvent.change(screen.getByLabelText(/^brand$/i), { target: { value: "Hot Wheels" } });
    fireEvent.click(screen.getByRole("button", { name: /add item/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          part_number: "PN-100",
          title: "Mainline",
          brand: "Hot Wheels",
        }),
      );
    });
  });
});
