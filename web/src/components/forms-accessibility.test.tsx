import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { axe } from "vitest-axe";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AIAssistForms } from "./ai-assist-forms";
import { CollectionItemForm } from "./collection-item-form";

describe("form accessibility", () => {
  afterEach(() => {
    cleanup();
  });

  it("associates validation errors and focuses first invalid field on submit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn(async () => undefined);
    render(<CollectionItemForm onSubmit={onSubmit} />);

    await user.click(screen.getByRole("button", { name: /add item/i }));
    expect(onSubmit).not.toHaveBeenCalled();

    const partNumber = screen.getByLabelText(/part number/i);
    await waitFor(() => expect(partNumber).toHaveFocus());
    expect(partNumber).toHaveAttribute("aria-invalid", "true");
    expect(partNumber).toHaveAttribute("aria-describedby", "part_number-error");
    expect(await screen.findByText(/part number is required/i)).toBeInTheDocument();
  });

  it("passes axe checks for AI assist form", async () => {
    const view = render(
      <AIAssistForms
        aiEnabled
        suggestion={{ title: "Suggested", confidence: 0.88 }}
        onToggle={() => undefined}
        onTest={() => undefined}
        onSuggestTitle={() => undefined}
        onSuggestPhoto={() => undefined}
        onApplySuggestion={() => undefined}
        onRetry={() => undefined}
      />,
    );
    const results = await axe(view.container, {
      rules: {
        "color-contrast": { enabled: false },
      },
    });
    expect(results).toHaveNoViolations();
  });

  it("supports keyboard-only confirmation and apply in AI assist flow", async () => {
    const user = userEvent.setup();
    const onApplySuggestion = vi.fn();
    render(
      <AIAssistForms
        aiEnabled
        suggestion={{ title: "Suggested", confidence: 0.88 }}
        onToggle={() => undefined}
        onTest={() => undefined}
        onSuggestTitle={() => undefined}
        onSuggestPhoto={() => undefined}
        onApplySuggestion={onApplySuggestion}
        onRetry={() => undefined}
      />,
    );

    const apply = screen.getByRole("button", { name: /apply suggestion/i });
    expect(apply).toBeDisabled();

    const confirm = screen.getByLabelText(/confirm apply suggestion/i);
    confirm.focus();
    expect(confirm).toHaveFocus();
    await user.keyboard("[Space]");
    expect(confirm).toBeChecked();

    await user.tab();
    expect(apply).toHaveFocus();
    await user.keyboard("[Enter]");
    expect(onApplySuggestion).toHaveBeenCalledTimes(1);
  });
});
