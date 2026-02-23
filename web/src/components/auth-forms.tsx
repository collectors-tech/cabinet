import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { BoundTextInput, Form, SubmitButton } from "./forms";

const recoverySchema = z.object({
  passphrase: z.string().trim().min(1, "Recovery passphrase is required"),
});

type RecoveryValues = z.infer<typeof recoverySchema>;

export function RecoveryPassphraseForm({
  onSubmit,
  isSubmitting,
}: {
  onSubmit: (values: RecoveryValues) => Promise<void> | void;
  isSubmitting?: boolean;
}) {
  const form = useForm<RecoveryValues>({
    resolver: zodResolver(recoverySchema),
    defaultValues: { passphrase: "" },
  });

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (values) => {
          await onSubmit(values);
        })}
      >
        <BoundTextInput<RecoveryValues, "passphrase">
          name="passphrase"
          label="Recovery passphrase"
          placeholder="Recovery passphrase"
        />
        <SubmitButton isSubmitting={Boolean(isSubmitting)}>
          Save Recovery Passphrase
        </SubmitButton>
      </form>
    </Form>
  );
}

const sessionSchema = z.object({
  token: z.string().trim().min(1, "Session token is required"),
});

type SessionValues = z.infer<typeof sessionSchema>;

export function SessionTokenForm({
  onValidate,
  onLock,
}: {
  onValidate: (values: SessionValues) => Promise<void> | void;
  onLock: (values: SessionValues) => Promise<void> | void;
}) {
  const form = useForm<SessionValues>({
    resolver: zodResolver(sessionSchema),
    defaultValues: { token: "" },
  });

  return (
    <Form {...form}>
      <form>
        <BoundTextInput<SessionValues, "token">
          name="token"
          label="Session token"
          placeholder="Session token"
        />
        <button
          type="button"
          onClick={form.handleSubmit(async (values) => {
            await onValidate(values);
          })}
        >
          Validate Session
        </button>{" "}
        <button
          type="button"
          onClick={form.handleSubmit(async (values) => {
            await onLock(values);
          })}
        >
          Lock Session
        </button>
      </form>
    </Form>
  );
}

