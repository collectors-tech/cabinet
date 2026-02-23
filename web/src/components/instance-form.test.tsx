import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { InstanceForm, type InstanceFormValues } from "./instance-form";

describe("InstanceForm", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders grouped instance fields and validates status options", async () => {
    const onSubmit = vi.fn(async (_values: InstanceFormValues) => {});
    render(<InstanceForm itemID="item-1" onSubmit={onSubmit} />);

    expect(screen.getByLabelText(/instance status/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/instance condition/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/quantity/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/instance status/i), { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: /add instance/i }));

    expect(await screen.findByText(/status is required/i)).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("submits valid instance payload", async () => {
    const onSubmit = vi.fn(async (_values: InstanceFormValues) => {});
    render(<InstanceForm itemID="item-1" onSubmit={onSubmit} />);

    fireEvent.change(screen.getByLabelText(/instance condition/i), { target: { value: "mint" } });
    fireEvent.change(screen.getByLabelText(/instance status/i), { target: { value: "sealed" } });
    fireEvent.change(screen.getByLabelText(/quantity/i), { target: { value: "2" } });
    fireEvent.click(screen.getByRole("button", { name: /add instance/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          item_id: "item-1",
          status: "sealed",
          condition: "mint",
          quantity: 2,
        }),
      );
    });
  });
});
