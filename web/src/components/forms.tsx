import * as React from "react";
import {
  Controller,
  FormProvider,
  useController,
  useFormContext,
  type ControllerProps,
  type FieldPath,
  type FieldValues,
  type UseFormReturn,
} from "react-hook-form";

type FormProps<TFieldValues extends FieldValues> = UseFormReturn<TFieldValues>;

const Form = <TFieldValues extends FieldValues>({
  children,
  ...form
}: FormProps<TFieldValues> & { children: React.ReactNode }) => (
  <FormProvider {...form}>{children}</FormProvider>
);

const FormField = <
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
>(
  props: ControllerProps<TFieldValues, TName>,
) => <Controller {...props} />;

type FormFieldContextValue = {
  name: string;
};

const FormFieldContext = React.createContext<FormFieldContextValue | null>(null);

const FormItem = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={["cabinet-form-item", className].filter(Boolean).join(" ")}
    {...props}
  />
));
FormItem.displayName = "FormItem";

const FormLabel = React.forwardRef<
  HTMLLabelElement,
  React.LabelHTMLAttributes<HTMLLabelElement>
>(({ className, ...props }, ref) => (
  <label
    ref={ref}
    className={["cabinet-form-label", className].filter(Boolean).join(" ")}
    {...props}
  />
));
FormLabel.displayName = "FormLabel";

const FormControl = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={["cabinet-form-control", className].filter(Boolean).join(" ")}
    {...props}
  />
));
FormControl.displayName = "FormControl";

function FormMessage({ className }: { className?: string }) {
  const { getFieldState, formState } = useFormContext();
  const field = React.useContext(FormFieldContext);
  if (!field) {
    return null;
  }
  const state = getFieldState(field.name, formState);
  if (!state.error?.message) {
    return null;
  }
  return (
    <p
      role="alert"
      className={["cabinet-form-message", className].filter(Boolean).join(" ")}
    >
      {String(state.error.message)}
    </p>
  );
}

type BoundFieldProps<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = {
  name: TName;
  label: string;
  placeholder?: string;
  type?: React.InputHTMLAttributes<HTMLInputElement>["type"];
};

function BoundTextInput<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
>({ name, label, placeholder, type = "text" }: BoundFieldProps<TFieldValues, TName>) {
  const { control } = useFormContext<TFieldValues>();
  const { field } = useController({ control, name });
  return (
    <div className="cabinet-form-item">
      <label className="cabinet-form-label" htmlFor={String(name)}>
        {label}
      </label>
      <input
        id={String(name)}
        className="cabinet-input"
        placeholder={placeholder}
        type={type}
        value={String(field.value ?? "")}
        onChange={field.onChange}
        onBlur={field.onBlur}
        name={field.name}
      />
      <FormFieldContext.Provider value={{ name: String(name) }}>
        <FormMessage />
      </FormFieldContext.Provider>
    </div>
  );
}

const TextInput = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    className={["cabinet-input", className].filter(Boolean).join(" ")}
    {...props}
  />
));
TextInput.displayName = "TextInput";

const TextAreaInput = React.forwardRef<
  HTMLTextAreaElement,
  React.TextareaHTMLAttributes<HTMLTextAreaElement>
>(({ className, ...props }, ref) => (
  <textarea
    ref={ref}
    className={["cabinet-textarea", className].filter(Boolean).join(" ")}
    {...props}
  />
));
TextAreaInput.displayName = "TextAreaInput";

const SelectInput = React.forwardRef<
  HTMLSelectElement,
  React.SelectHTMLAttributes<HTMLSelectElement>
>(({ className, ...props }, ref) => (
  <select
    ref={ref}
    className={["cabinet-select", className].filter(Boolean).join(" ")}
    {...props}
  />
));
SelectInput.displayName = "SelectInput";

function MultiSelectInput({
  options,
  value,
  onChange,
  ...props
}: {
  options: Array<{ label: string; value: string }>;
  value: string[];
  onChange: (next: string[]) => void;
} & Omit<React.SelectHTMLAttributes<HTMLSelectElement>, "value" | "onChange" | "multiple">) {
  return (
    <select
      {...props}
      multiple
      className={["cabinet-select", props.className].filter(Boolean).join(" ")}
      value={value}
      onChange={(event) => {
        const next = Array.from(event.currentTarget.selectedOptions).map((opt) => opt.value);
        onChange(next);
      }}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

const CheckboxInput = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    type="checkbox"
    className={["cabinet-checkbox", className].filter(Boolean).join(" ")}
    {...props}
  />
));
CheckboxInput.displayName = "CheckboxInput";

function SwitchInput({
  checked,
  onChange,
  className,
  ...props
}: {
  checked: boolean;
  onChange: (checked: boolean) => void;
} & Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, "onChange">) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className={[
        "cabinet-switch",
        checked ? "cabinet-switch-on" : "cabinet-switch-off",
        className,
      ]
        .filter(Boolean)
        .join(" ")}
      {...props}
    />
  );
}

const DateInput = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    type="date"
    className={["cabinet-input", className].filter(Boolean).join(" ")}
    {...props}
  />
));
DateInput.displayName = "DateInput";

const CurrencyInput = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, inputMode, ...props }, ref) => (
  <input
    ref={ref}
    type="text"
    inputMode={inputMode || "decimal"}
    className={["cabinet-input", className].filter(Boolean).join(" ")}
    {...props}
  />
));
CurrencyInput.displayName = "CurrencyInput";

function SubmitButton({
  children,
  isSubmitting,
  disabled,
}: {
  children: React.ReactNode;
  isSubmitting: boolean;
  disabled?: boolean;
}) {
  return (
    <button type="submit" disabled={disabled || isSubmitting}>
      {isSubmitting ? "Saving..." : children}
    </button>
  );
}

export {
  BoundTextInput,
  CheckboxInput,
  CurrencyInput,
  DateInput,
  Form,
  FormControl,
  FormField,
  FormFieldContext,
  FormItem,
  FormLabel,
  FormMessage,
  MultiSelectInput,
  SelectInput,
  SubmitButton,
  SwitchInput,
  TextAreaInput,
  TextInput,
};
