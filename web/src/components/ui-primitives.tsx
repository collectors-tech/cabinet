import { useEffect, useId, useRef } from "react";

type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";
type ButtonSize = "sm" | "md" | "lg";

function Button({
  variant,
  size,
  loading,
  disabled,
  className,
  children,
  onClick,
  ...props
}: {
  variant: ButtonVariant;
  size: ButtonSize;
  loading?: boolean;
  disabled?: boolean;
  className?: string;
  children: React.ReactNode;
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
} & Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, "onClick" | "disabled">) {
  const blocked = Boolean(loading || disabled);
  return (
    <button
      type="button"
      {...props}
      className={["cabinet-btn", `cabinet-btn-${variant}`, `cabinet-btn-${size}`, className].filter(Boolean).join(" ")}
      disabled={blocked}
      aria-busy={loading ? "true" : undefined}
      onClick={(event) => {
        if (blocked) {
          event.preventDefault();
          return;
        }
        onClick?.(event);
      }}
    >
      {loading ? "Saving..." : children}
    </button>
  );
}

function TextField({
  value,
  onChange,
  placeholder,
  invalid,
  message,
  id,
  ...props
}: {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  invalid?: boolean;
  message?: string;
  id?: string;
} & Omit<React.InputHTMLAttributes<HTMLInputElement>, "value" | "onChange">) {
  const fallbackID = useId();
  const inputID = id || fallbackID;
  const messageID = `${inputID}-message`;
  return (
    <div className="cabinet-field">
      <input
        {...props}
        id={inputID}
        className={["cabinet-input", invalid ? "cabinet-input-invalid" : ""].filter(Boolean).join(" ")}
        value={value}
        onChange={(event) => onChange(event.currentTarget.value)}
        placeholder={placeholder}
        aria-invalid={invalid ? "true" : undefined}
        aria-describedby={invalid && message ? messageID : undefined}
      />
      {invalid && message ? (
        <p id={messageID} className="cabinet-field-error" role="alert">
          {message}
        </p>
      ) : null}
    </div>
  );
}

function SelectField({
  value,
  options,
  onChange,
  ...props
}: {
  value: string;
  options: Array<{ value: string; label: string }>;
  onChange: (value: string) => void;
} & Omit<React.SelectHTMLAttributes<HTMLSelectElement>, "value" | "onChange">) {
  return (
    <select {...props} className={["cabinet-select", props.className].filter(Boolean).join(" ")} value={value} onChange={(event) => onChange(event.currentTarget.value)}>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

function Dialog({
  open,
  onOpenChange,
  title,
  description,
  triggerSelector,
  children,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  triggerSelector?: string;
  children: React.ReactNode;
}) {
  const contentRef = useRef<HTMLDivElement | null>(null);
  const restoreRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    restoreRef.current = triggerSelector ? (document.querySelector(triggerSelector) as HTMLElement | null) : (document.activeElement as HTMLElement | null);
    contentRef.current?.focus();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onOpenChange(false);
        restoreRef.current?.focus();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open, onOpenChange, triggerSelector]);

  if (!open) {
    return null;
  }
  return (
    <div className="cabinet-dialog-overlay" onClick={() => onOpenChange(false)}>
      <div
        ref={contentRef}
        tabIndex={-1}
        role="dialog"
        aria-modal="true"
        aria-label={title}
        className="cabinet-dialog"
        onClick={(event) => event.stopPropagation()}
      >
        <h3>{title}</h3>
        {description ? <p>{description}</p> : null}
        <button type="button" aria-label="Close dialog" onClick={() => onOpenChange(false)}>
          Close
        </button>
        {children}
      </div>
    </div>
  );
}

type DrawerItem = {
  id: string;
  label: string;
  route: string;
};

function Drawer({
  open,
  onClose,
  items,
  onNavigate,
}: {
  open: boolean;
  onClose: () => void;
  items: DrawerItem[];
  onNavigate: (route: string) => void;
}) {
  const drawerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    drawerRef.current?.focus();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open, onClose]);

  if (!open) {
    return null;
  }

  return (
    <div className="cabinet-drawer-overlay" onClick={onClose}>
      <div
        ref={drawerRef}
        tabIndex={-1}
        role="dialog"
        aria-modal="true"
        aria-label="Navigation Drawer"
        className="cabinet-drawer-panel"
        onClick={(event) => event.stopPropagation()}
      >
        <button type="button" onClick={onClose} aria-label="Close drawer">
          Close
        </button>
        <nav aria-label="Drawer navigation">
          {items.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => {
                onNavigate(item.route);
                onClose();
              }}
            >
              {item.label}
            </button>
          ))}
        </nav>
      </div>
    </div>
  );
}

export { Button, Dialog, Drawer, SelectField, TextField };
