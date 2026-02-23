# Form Architecture

Cabinet uses `react-hook-form` + `zod` as the standard form stack for the UI.

## Primitives

Shared form primitives live in `web/src/components/forms.tsx`:

- `Form`
- `FormField`
- `FormLabel`
- `FormControl`
- `FormMessage`
- `SubmitButton`

Reusable inputs in the same file:

- `TextInput`
- `TextAreaInput`
- `SelectInput`
- `MultiSelectInput`
- `CheckboxInput`
- `SwitchInput`
- `DateInput`
- `CurrencyInput`

## Validation Pattern

- Each form declares a `zod` schema.
- `zodResolver` is used in `useForm`.
- Field-level errors are surfaced via `FormMessage`.
- Submit buttons use a single loading/disabled pattern via `SubmitButton`.

## Production Usage

`web/src/components/starter-quick-add-form.tsx` is the first production form migrated to this stack and is wired in `web/src/App.tsx` for starter onboarding.
