import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AIAssistForms } from "./ai-assist-forms";

describe("AIAssistForms", () => {
  afterEach(() => {
    cleanup();
  });

  it("shows title normalization flow with explicit confirmation gate before apply", async () => {
    const onSuggestTitle = vi.fn(async () => undefined);
    const onSuggestPhoto = vi.fn(async () => undefined);
    const onApplySuggestion = vi.fn();
    const onRetry = vi.fn(async () => undefined);

    render(
      <AIAssistForms
        aiEnabled
        suggestion={{ title: "Suggested Title", confidence: 0.91 }}
        onToggle={async () => undefined}
        onTest={async () => undefined}
        onSuggestTitle={onSuggestTitle}
        onSuggestPhoto={onSuggestPhoto}
        onApplySuggestion={onApplySuggestion}
        onRetry={onRetry}
      />,
    );

    fireEvent.change(screen.getByLabelText(/title normalization input/i), { target: { value: "raw listing title" } });
    fireEvent.click(screen.getByRole("button", { name: /normalize title/i }));
    await waitFor(() => expect(onSuggestTitle).toHaveBeenCalled());

    fireEvent.click(screen.getByRole("button", { name: /apply suggestion/i }));
    expect(onApplySuggestion).not.toHaveBeenCalled();

    fireEvent.click(screen.getByLabelText(/confirm apply suggestion/i));
    fireEvent.click(screen.getByRole("button", { name: /apply suggestion/i }));
    expect(onApplySuggestion).toHaveBeenCalledTimes(1);
  });

  it("shows photo identify preview with retry action when AI fails", async () => {
    const onSuggestTitle = vi.fn(async () => undefined);
    const onSuggestPhoto = vi.fn(async () => undefined);
    const onRetry = vi.fn(async () => undefined);

    render(
      <AIAssistForms
        aiEnabled
        aiError="failed_ai_suggest_photo"
        lastAction="photo"
        onToggle={async () => undefined}
        onTest={async () => undefined}
        onSuggestTitle={onSuggestTitle}
        onSuggestPhoto={onSuggestPhoto}
        onApplySuggestion={() => undefined}
        onRetry={onRetry}
      />,
    );

    fireEvent.change(screen.getByLabelText(/photo identify url/i), { target: { value: "https://example.com/item.jpg" } });
    expect(screen.getByRole("img", { name: /photo preview/i })).toHaveAttribute("src", "https://example.com/item.jpg");
    fireEvent.click(screen.getByRole("button", { name: /identify from photo/i }));
    await waitFor(() => expect(onSuggestPhoto).toHaveBeenCalled());
    fireEvent.click(screen.getByRole("button", { name: /retry last ai action/i }));
    await waitFor(() => expect(onRetry).toHaveBeenCalled());
  });
});
