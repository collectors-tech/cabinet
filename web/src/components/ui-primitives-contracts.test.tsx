import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { Button, Dialog, Drawer, SelectField, TextField } from "./ui-primitives";

describe("strict primitive contracts", () => {
  afterEach(() => {
    cleanup();
  });

  it("BTN-001 disabled blocks click", () => {
    const onClick = vi.fn();
    render(
      <Button variant="primary" size="md" disabled onClick={onClick}>
        Save
      </Button>,
    );
    fireEvent.click(screen.getByRole("button", { name: /save/i }));
    expect(onClick).not.toHaveBeenCalled();
  });

  it("BTN-002 loading blocks repeat submits", () => {
    const onClick = vi.fn();
    render(
      <Button variant="primary" size="md" loading onClick={onClick}>
        Save
      </Button>,
    );
    const button = screen.getByRole("button", { name: /saving/i });
    fireEvent.click(button);
    fireEvent.click(button);
    expect(onClick).not.toHaveBeenCalled();
    expect(button).toBeDisabled();
  });

  it("INP-001 invalid state shows message", () => {
    render(
      <TextField
        value="PN"
        onChange={() => undefined}
        placeholder="Part number"
        aria-label="Part Number"
        invalid
        message="Part number is required"
      />,
    );
    expect(screen.getByRole("alert")).toHaveTextContent(/part number is required/i);
  });

  it("SEL-001 option change updates bound state", () => {
    const onChange = vi.fn();
    render(
      <SelectField
        value="one"
        options={[
          { value: "one", label: "One" },
          { value: "two", label: "Two" },
        ]}
        onChange={onChange}
        aria-label="Select demo"
      />,
    );
    fireEvent.change(screen.getByLabelText(/select demo/i), { target: { value: "two" } });
    expect(onChange).toHaveBeenCalledWith("two");
  });

  it("DIA-001 escape closes dialog", () => {
    const onOpenChange = vi.fn();
    render(
      <Dialog open onOpenChange={onOpenChange} title="Dialog title" description="Dialog description">
        <button type="button">Inside</button>
      </Dialog>,
    );
    fireEvent.keyDown(document, { key: "Escape" });
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("DIA-002 focus returns to trigger", () => {
    const onOpenChange = vi.fn();
    render(
      <>
        <button type="button" data-testid="trigger">
          Open dialog
        </button>
        <Dialog open onOpenChange={onOpenChange} title="Dialog title" triggerSelector='[data-testid="trigger"]'>
          <button type="button">Inside</button>
        </Dialog>
      </>,
    );
    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.getByTestId("trigger")).toHaveFocus();
  });

  it("DRW-001 opens/closes via trigger and escape", () => {
    const onClose = vi.fn();
    render(
      <Drawer
        open
        onClose={onClose}
        items={[
          { id: "dashboard", label: "Dashboard", route: "dashboard" },
          { id: "collection", label: "Collection", route: "collection" },
        ]}
        onNavigate={() => undefined}
      />,
    );
    expect(screen.getByRole("dialog", { name: /navigation drawer/i })).toBeInTheDocument();
    fireEvent.keyDown(document, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });

  it("DRW-002 navigation closes drawer", () => {
    const onClose = vi.fn();
    const onNavigate = vi.fn();
    render(
      <Drawer
        open
        onClose={onClose}
        items={[
          { id: "dashboard", label: "Dashboard", route: "dashboard" },
          { id: "collection", label: "Collection", route: "collection" },
        ]}
        onNavigate={onNavigate}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /collection/i }));
    expect(onNavigate).toHaveBeenCalledWith("collection");
    expect(onClose).toHaveBeenCalled();
  });
});
