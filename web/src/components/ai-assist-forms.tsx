import { useMemo, useState } from "react";

type Suggestion = {
  title?: string;
  confidence?: number;
  [key: string]: unknown;
};

export function AIAssistForms({
  aiEnabled,
  aiError,
  suggestion,
  lastAction,
  onToggle,
  onTest,
  onSuggestTitle,
  onSuggestPhoto,
  onApplySuggestion,
  onRetry,
}: {
  aiEnabled: boolean;
  aiError?: string;
  suggestion?: Suggestion | null;
  lastAction?: "title" | "photo" | "";
  onToggle: (enabled: boolean) => Promise<void> | void;
  onTest: () => Promise<void> | void;
  onSuggestTitle: (title: string) => Promise<void> | void;
  onSuggestPhoto: (photoURL: string) => Promise<void> | void;
  onApplySuggestion: () => void;
  onRetry: () => Promise<void> | void;
}) {
  const [titleInput, setTitleInput] = useState("");
  const [photoURL, setPhotoURL] = useState("");
  const [confirmApply, setConfirmApply] = useState(false);

  const hasPhotoPreview = useMemo(() => /^https?:\/\/\S+/i.test(photoURL.trim()), [photoURL]);

  return (
    <div>
      <h4>AI Assist Workflows</h4>
      <div>
        <button type="button" onClick={() => onToggle(true)}>
          Enable AI
        </button>{" "}
        <button type="button" onClick={() => onToggle(false)}>
          Disable AI
        </button>{" "}
        <button type="button" onClick={onTest}>
          Test AI
        </button>
      </div>
      <p>AI enabled: {String(aiEnabled)}</p>

      <div>
        <label htmlFor="ai-title-normalize">Title normalization input</label>
        <input
          id="ai-title-normalize"
          aria-label="Title normalization input"
          value={titleInput}
          placeholder="Listing title"
          onChange={(e) => setTitleInput(e.target.value)}
        />{" "}
        <button type="button" onClick={() => onSuggestTitle(titleInput.trim())}>
          Normalize Title
        </button>
      </div>

      <div>
        <label htmlFor="ai-photo-identify">Photo identify URL</label>
        <input
          id="ai-photo-identify"
          aria-label="Photo identify URL"
          value={photoURL}
          placeholder="Image URL"
          onChange={(e) => setPhotoURL(e.target.value)}
        />{" "}
        <button type="button" onClick={() => onSuggestPhoto(photoURL.trim())}>
          Identify From Photo
        </button>
      </div>
      {hasPhotoPreview ? <img alt="Photo preview" src={photoURL.trim()} style={{ maxWidth: "160px", maxHeight: "120px" }} /> : null}

      {suggestion ? (
        <div>
          <p>AI confidence: {String(suggestion.confidence ?? "")}</p>
          <p>AI title: {String(suggestion.title ?? "")}</p>
          <label>
            <input
              type="checkbox"
              aria-label="Confirm apply suggestion"
              checked={confirmApply}
              onChange={(e) => setConfirmApply(e.target.checked)}
            />
            Confirm apply suggestion
          </label>{" "}
          <button type="button" disabled={!confirmApply} onClick={onApplySuggestion}>
            Apply Suggestion
          </button>
          <p>Records are never auto-created. Explicit confirmation is required.</p>
        </div>
      ) : null}
      {aiError ? <p>AI error: {aiError}</p> : null}
      {aiError && lastAction ? (
        <button type="button" onClick={onRetry}>
          Retry Last AI Action
        </button>
      ) : null}
    </div>
  );
}
